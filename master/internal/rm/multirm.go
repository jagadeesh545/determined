package rm

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/determined-ai/determined/master/internal/config"
	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/internal/rm/agentrm"
	"github.com/determined-ai/determined/master/internal/rm/kubernetesrm"
	"github.com/determined-ai/determined/master/internal/sproto"
	"github.com/determined-ai/determined/master/pkg/aproto"
	"github.com/determined-ai/determined/master/pkg/command"
	"github.com/determined-ai/determined/master/pkg/model"
	"github.com/determined-ai/determined/proto/pkg/agentv1"
	"github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/jobv1"
	"github.com/determined-ai/determined/proto/pkg/resourcepoolv1"
)

// ErrRPNotDefined returns a detailed error if a resource manager/pool combination isn't found in the MultiRMRouter map.
func ErrRPNotDefined(rm, rp string) error {
	return fmt.Errorf("resource pool %s not found in resource manager %s", rp, rm)
}

// MultiRMRouter tracks all resource managers in the system.
type MultiRMRouter struct {
	mu     sync.Mutex
	rms    map[string]ResourceManager
	syslog *logrus.Entry
}

// RMRouter TODO (multirm) -- GetResourcePools needs its own interface to get around
// an import cycle in telemetry module.
type RMRouter interface {
	GetResourcePools() (*apiv1.GetResourcePoolsResponse, error)
}

// DefaultRMRouter is the global MultiRMRouter.
var DefaultRMRouter *MultiRMRouter

// SetDefaultRouter initializes DefaultRMRouter with the master set-up.
func SetDefaultRouter(
	db *db.PgDB,
	echo *echo.Echo,
	rmConfigs []config.ResourceManagerWithPoolsConfig,
	tcd *model.TaskContainerDefaultsConfig,
	opts *aproto.MasterSetAgentOptions,
	cert *tls.Certificate,
) MultiRMRouter {
	rms := map[string]ResourceManager{}
	for _, cfg := range rmConfigs {
		c := cfg
		switch {
		case c.ResourceManager.AgentRM != nil:
			// TODO (multirm): what to do if Name is ""?
			rms[c.ResourceManager.AgentRM.Name] = agentrm.New(db, echo, &c, opts, cert)
		case c.ResourceManager.KubernetesRM != nil:
			// TODO (multirm): ditto.
			rms[c.ResourceManager.KubernetesRM.Name] = kubernetesrm.New(db, &c, tcd, opts, cert)
		default:
			panic("no expected resource manager config is defined")
		}
	}
	return MultiRMRouter{
		rms:    rms,
		syslog: logrus.WithField("component", "resource-router"),
	}
}

// GetAllocationSummaries returns the allocation summaries for all resource pools across all resource managers.
func GetAllocationSummaries() (
	map[model.AllocationID]sproto.AllocationSummary,
	error,
) {
	DefaultRMRouter.mu.Lock()
	defer DefaultRMRouter.mu.Unlock()

	var eg errgroup.Group
	var mu sync.Mutex
	res := map[model.AllocationID]sproto.AllocationSummary{}

	for _, rm := range DefaultRMRouter.rms {
		r := rm
		eg.Go(func() error {
			res1, err := r.GetAllocationSummaries()
			if err != nil {
				return err
			}

			// TODO (multirm): I know this looks ugly, but I couldn't
			// figure out how to *not* panic if I send this to a channel.
			mu.Lock()
			// Note: this is a built-in function that will rewrite on duplicates.
			maps.Copy(res, res1)
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

// Allocate routes an AllocateRequest to the specified RM.
func Allocate(req sproto.AllocateRequest) (*sproto.ResourcesSubscription, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		return nil, ErrRPNotDefined(req.ResourceManager, req.ResourcePool)
	}
	return rmForRp.Allocate(req)
}

// Release routes an allocation release request.
func Release(req sproto.AllocateRequest, resourceID *sproto.ResourcesID) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.ResourcePool))
		return
	}

	rmForRp.Release(sproto.ResourcesReleased{
		AllocationID: req.AllocationID,
		ResourcesID:  resourceID,
		ResourcePool: req.ResourcePool,
	})
}

// ValidateResources routes a validation request for a specified resource manager/pool.
func ValidateResources(rm, rp string, slots int, command bool) error {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return ErrRPNotDefined(rm, rp)
	}
	return rmForRp.ValidateResources(rp, slots, command)
}

// DeleteJob routes a DeleteJob request to the specified resource manager.
func DeleteJob(rm string, jobID model.JobID) (sproto.DeleteJobResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(rm, ""))
	}
	return rmForRp.DeleteJob(jobID)
}

// NotifyContainerRunning routes a NotifyContainerRunning request to a specified resource manager/pool.
func NotifyContainerRunning(rm string, req sproto.NotifyContainerRunning) error {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(rm, ""))
	}
	return rmForRp.NotifyContainerRunning(req)
}

// SetGroupMaxSlots routes a SetGroupMaxSlots request to a specified resource manager/pool.
func SetGroupMaxSlots(rm string, req sproto.SetGroupMaxSlots) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(rm, req.ResourcePool))
	}
	rmForRp.SetGroupMaxSlots(req)
}

// SetGroupWeight routes a SetGroupWeight request to a specified resource manager/pool.
func SetGroupWeight(rm string, req sproto.SetGroupWeight) error {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return ErrRPNotDefined(rm, req.ResourcePool)
	}
	return rmForRp.SetGroupWeight(req)
}

// SetGroupPriority routes a SetGroupPriority request to a specified resource manager/pool.
func SetGroupPriority(rm string, req sproto.SetGroupPriority) error {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return ErrRPNotDefined(rm, req.ResourcePool)
	}
	return rmForRp.SetGroupPriority(req)
}

// ExternalPreemptionPending routes an ExternalPreemptionPending request to the specified resource manager.
func ExternalPreemptionPending(rm string, allocationID model.AllocationID) error {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return ErrRPNotDefined(rm, "")
	}
	return rmForRp.ExternalPreemptionPending(allocationID)
}

// IsReattachableOnlyAfterStarted routes a IsReattachableOnlyAfterStarted call to a specified resource manager/pool.
func IsReattachableOnlyAfterStarted(rm string) bool {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return false // TODO (multirm): how else to log this?
	}
	return rmForRp.IsReattachableOnlyAfterStarted()
}

// GetResourcePools returns all resource pools across all resource managers.
func (r *MultiRMRouter) GetResourcePools() (*apiv1.GetResourcePoolsResponse, error) {
	res := make(chan []*resourcepoolv1.ResourcePool)

	// TODO (multirm): switch this to errgroup?
	var wg sync.WaitGroup
	wg.Add(len(r.rms))

	for _, rm := range r.rms {
		go func(m ResourceManager) {
			defer wg.Done()
			res1, err := m.GetResourcePools()
			if err != nil {
				r.syslog.WithError(err)
			}
			res <- res1.ResourcePools
		}(rm)
	}
	wg.Wait()

	close(res)

	var totalRes []*resourcepoolv1.ResourcePool
	for x := range res {
		totalRes = append(totalRes, x...)
	}
	return &apiv1.GetResourcePoolsResponse{ResourcePools: totalRes}, nil
}

// GetDefaultComputeResourcePool routes a GetDefaultComputeResourcePool to the specified resource manager.
func GetDefaultComputeResourcePool(rm string) (sproto.GetDefaultComputeResourcePoolResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return sproto.GetDefaultComputeResourcePoolResponse{}, ErrRPNotDefined(rm, "")
	}
	return rmForRp.GetDefaultComputeResourcePool()
}

// GetDefaultAuxResourcePool routes a GetDefaultAuxResourcePool to the specified resource manager.
func GetDefaultAuxResourcePool(rm string) (sproto.GetDefaultAuxResourcePoolResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return sproto.GetDefaultAuxResourcePoolResponse{}, ErrRPNotDefined(rm, "")
	}
	return rmForRp.GetDefaultAuxResourcePool()
}

// ResolveResourcePool routes a ResolveResourcePool request for a specific resource manager/pool.
func ResolveResourcePool(rm, rp string, workspace, slots int) (resourcePool string, err error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return rp, ErrRPNotDefined(rm, rp)
	}
	return rmForRp.ResolveResourcePool(rp, workspace, slots)
}

// ValidateResourcePoolAvailability routes a ValidateResourcePoolAvailability call for a specific RM/RP.
func ValidateResourcePoolAvailability(
	rm string,
	v *sproto.ValidateResourcePoolAvailabilityRequest,
) ([]command.LaunchWarning, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return []command.LaunchWarning{}, ErrRPNotDefined(rm, v.Name)
	}
	return rmForRp.ValidateResourcePoolAvailability(v)
}

// TaskContainerDefaults routes a TaskContainerDefaults call to a specific resource manager/pool.
func TaskContainerDefaults(
	rm, rp string,
	fallbackConfig model.TaskContainerDefaultsConfig,
) (model.TaskContainerDefaultsConfig, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return model.TaskContainerDefaultsConfig{}, ErrRPNotDefined(rm, rp)
	}
	return rmForRp.TaskContainerDefaults(rp, fallbackConfig)
}

// GetJobQ routes a GetJobQ call to a specified resource manager/pool.
func GetJobQ(rm, rp string) (map[model.JobID]*sproto.RMJobInfo, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return nil, ErrRPNotDefined(rm, rp)
	}
	return rmForRp.GetJobQ(rp)
}

// GetJobQueueStatsRequest routes a GetJobQueueStatsRequest to the specified resource manager.
func GetJobQueueStatsRequest(rm string, req *apiv1.GetJobQueueStatsRequest) (*apiv1.GetJobQueueStatsResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return nil, ErrRPNotDefined(rm, strings.Join(req.ResourcePools, ", "))
	}
	return rmForRp.GetJobQueueStatsRequest(req)
}

// MoveJob routes a MoveJob call to a specified resource manager/pool.
func MoveJob(rm string, req sproto.MoveJob) error {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return ErrRPNotDefined(rm, req.ResourcePool)
	}
	return rmForRp.MoveJob(req)
}

// RecoverJobPosition routes a RecoverJobPosition call to a specified resource manager/pool.
func RecoverJobPosition(rm string, req sproto.RecoverJobPosition) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(rm, req.ResourcePool))
	}
	rmForRp.RecoverJobPosition(req)
}

// GetExternalJobs routes a GetExternalJobs request to a specified resource manager.
func GetExternalJobs(rm string, rp string) ([]*jobv1.Job, error) {
	rmForRp, ok := DefaultRMRouter.rms[rm]
	if !ok {
		return nil, ErrRPNotDefined(rm, "")
	}
	return rmForRp.GetExternalJobs(rp)
}

// GetAgents returns all agents across all resource managers.
func GetAgents() (*apiv1.GetAgentsResponse, error) {
	res := make(chan []*agentv1.Agent)

	// TODO (multirm): switch this to errgroup?
	var wg sync.WaitGroup
	wg.Add(len(DefaultRMRouter.rms))

	for _, rm := range DefaultRMRouter.rms {
		go func(r ResourceManager) {
			defer wg.Done()
			res1, err := r.GetAgents()
			if err != nil {
				DefaultRMRouter.syslog.WithError(err)
			}
			res <- res1.Agents
		}(rm)
	}
	wg.Wait()

	close(res)

	var totalRes []*agentv1.Agent
	for r := range res {
		totalRes = append(totalRes, r...)
	}
	return &apiv1.GetAgentsResponse{Agents: totalRes}, nil
}

// GetAgent routes a GetAgent request to the specified resource manager & agent.
func GetAgent(req *apiv1.GetAgentRequest) (*apiv1.GetAgentResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.GetAgent(req)
}

// EnableAgent routes an EnableAgent request to the specified resource manager & agent.
func EnableAgent(req *apiv1.EnableAgentRequest) (*apiv1.EnableAgentResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.EnableAgent(req)
}

// DisableAgent routes an DisableAgent request to the specified resource manager & agent.
func DisableAgent(req *apiv1.DisableAgentRequest) (*apiv1.DisableAgentResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.DisableAgent(req)
}

// GetSlots routes an GetSlots request to the specified resource manager & agent.
func GetSlots(req *apiv1.GetSlotsRequest) (*apiv1.GetSlotsResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.GetSlots(req)
}

// GetSlot routes an GetSlot request to the specified resource manager & agent.
func GetSlot(req *apiv1.GetSlotRequest) (*apiv1.GetSlotResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.GetSlot(req)
}

// EnableSlot routes an EnableSlot request to the specified resource manager & agent.
func EnableSlot(req *apiv1.EnableSlotRequest) (*apiv1.EnableSlotResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.EnableSlot(req)
}

// DisableSlot routes an DisableSlot request to the specified resource manager & agent.
func DisableSlot(req *apiv1.DisableSlotRequest) (*apiv1.DisableSlotResponse, error) {
	rmForRp, ok := DefaultRMRouter.rms[req.ResourceManager]
	if !ok {
		DefaultRMRouter.syslog.WithError(ErrRPNotDefined(req.ResourceManager, req.AgentId))
	}
	return rmForRp.DisableSlot(req)
}
