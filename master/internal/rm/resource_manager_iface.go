package rm

import (
	"github.com/determined-ai/determined/master/internal/sproto"
	"github.com/determined-ai/determined/master/pkg/command"
	"github.com/determined-ai/determined/master/pkg/model"
	"github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/jobv1"
)

// ResourceManager is an interface for a resource manager, which can allocate and manage resources.
type ResourceManager interface {
	// Basic functionality
	GetAllocationSummaries() (map[model.AllocationID]sproto.AllocationSummary, error)
	Allocate(sproto.AllocateRequest) (*sproto.ResourcesSubscription, error)
	Release(sproto.ResourcesReleased)
	ValidateResources(sproto.ValidateResources) error
	DeleteJob(rmName string, jobID model.JobID) (sproto.DeleteJobResponse, error)
	NotifyContainerRunning(sproto.NotifyContainerRunning) error

	// Scheduling related stuff
	SetGroupMaxSlots(sproto.SetGroupMaxSlots)
	SetGroupWeight(sproto.SetGroupWeight) error
	SetGroupPriority(sproto.SetGroupPriority) error
	ExternalPreemptionPending(rmName string, allocID model.AllocationID) error
	IsReattachableOnlyAfterStarted(rmName string) bool

	// Resource pool stuff
	GetResourcePools() (*apiv1.GetResourcePoolsResponse, error)
	GetDefaultComputeResourcePool() (sproto.GetDefaultComputeResourcePoolResponse, error)
	GetDefaultAuxResourcePool() (sproto.GetDefaultAuxResourcePoolResponse, error)
	ValidateResourcePool(rmName string, rpName string) error
	ResolveResourcePool(sproto.ResolveResourcesRequest) (string, string, error)
	ValidateResourcePoolAvailability(v *sproto.ValidateResourcePoolAvailabilityRequest) ([]command.LaunchWarning, error)
	TaskContainerDefaults(rmName string, rpName string, fallbackConfig model.TaskContainerDefaultsConfig) (
		model.TaskContainerDefaultsConfig, error)

	// Job queue
	GetJobQ(rmName string, rpName string) (map[model.JobID]*sproto.RMJobInfo, error)
	GetJobQueueStatsRequest(*apiv1.GetJobQueueStatsRequest) (*apiv1.GetJobQueueStatsResponse, error)
	MoveJob(sproto.MoveJob) error
	RecoverJobPosition(sproto.RecoverJobPosition)
	GetExternalJobs(rmName string, rpName string) ([]*jobv1.Job, error)

	// Cluster Management APIs
	GetAgents() (*apiv1.GetAgentsResponse, error)
	GetAgent(*apiv1.GetAgentRequest) (*apiv1.GetAgentResponse, error)
	EnableAgent(*apiv1.EnableAgentRequest) (*apiv1.EnableAgentResponse, error)
	DisableAgent(*apiv1.DisableAgentRequest) (*apiv1.DisableAgentResponse, error)
	GetSlots(*apiv1.GetSlotsRequest) (*apiv1.GetSlotsResponse, error)
	GetSlot(*apiv1.GetSlotRequest) (*apiv1.GetSlotResponse, error)
	EnableSlot(*apiv1.EnableSlotRequest) (*apiv1.EnableSlotResponse, error)
	DisableSlot(*apiv1.DisableSlotRequest) (*apiv1.DisableSlotResponse, error)
}
