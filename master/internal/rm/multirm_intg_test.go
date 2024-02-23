//go:build integration
// +build integration

package rm

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/determined-ai/determined/master/internal/config"
	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/internal/mocks"
	"github.com/determined-ai/determined/master/internal/sproto"
	"github.com/determined-ai/determined/master/pkg/aproto"
	"github.com/determined-ai/determined/master/pkg/command"
	"github.com/determined-ai/determined/master/pkg/etc"
	"github.com/determined-ai/determined/master/pkg/model"
	"github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/resourcepoolv1"
)

var pgDB *db.PgDB

// TestMain sets up the DB for tests.
func TestMain(m *testing.M) {
	tmp, err := db.ResolveTestPostgres()
	pgDB = tmp
	if err != nil {
		log.Panicln(err)
	}

	err = db.MigrateTestPostgres(pgDB, "file://../../static/migrations", "up")
	if err != nil {
		log.Panicln(err)
	}

	err = etc.SetRootPath("../../static/srv")
	if err != nil {
		log.Panicln(err)
	}

	os.Exit(m.Run())
}

func TestSetDefaultRouter(t *testing.T) {
	// TODO : I can't seem to mock out k8s RM.
	var cfg []config.ResourceManagerWithPoolsConfig
	var names []string
	for i := 1; i <= 3; i++ {
		c := mockRMConfig(2)
		cfg = append(cfg, c)
		names = append(names, c.ResourceManager.Name())
	}

	newRouter := SetDefaultRouter(pgDB, echo.New(), cfg, nil,
		&aproto.MasterSetAgentOptions{LoggingOptions: model.LoggingConfig{}}, nil)
	require.Equal(t, 3, len(newRouter.rms))
	for _, n := range names {
		require.NotNil(t, newRouter.rms[n])
	}
}

func TestGetAllocationSummaries(t *testing.T) {
	cases := []struct {
		name       string
		allocNames []string
		managers   int
	}{
		{"simple", []string{uuid.NewString(), uuid.NewString()}, 1},
		{"multirm", []string{uuid.NewString(), uuid.NewString()}, 3},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rms := map[string]ResourceManager{}
			for i := 1; i <= tt.managers; i++ {
				ret := map[model.AllocationID]sproto.AllocationSummary{}
				for _, alloc := range tt.allocNames {
					a := alloc + fmt.Sprint(i)
					ret[*model.NewAllocationID(&a)] = sproto.AllocationSummary{}
				}
				require.Equal(t, len(ret), len(tt.allocNames))

				mockRM := mocks.ResourceManager{}
				mockRM.On("GetAllocationSummaries").Return(ret, nil)

				rms[uuid.NewString()] = &mockRM
			}

			DefaultRMRouter = &MultiRMRouter{rms: rms}

			allocs, err := DefaultRMRouter.GetAllocationSummaries()
			require.NoError(t, err)
			require.Equal(t, tt.managers*len(tt.allocNames), len(allocs))
			require.NotNil(t, allocs)

			bogus := "bogus"
			require.Empty(t, allocs[*model.NewAllocationID(&bogus)])

			for _, name := range tt.allocNames {
				n := fmt.Sprintf(name + "0")
				require.NotNil(t, allocs[*model.NewAllocationID(&n)])
				require.Empty(t, allocs[*model.NewAllocationID(&name)])
			}
		})
	}
}

func TestAllocate(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	allocReq := sproto.AllocateRequest{ResourceManager: manager}
	mockRM.On("Allocate", allocReq).Return(&sproto.ResourcesSubscription{}, nil)

	res, err := DefaultRMRouter.Allocate(allocReq)
	require.NoError(t, err)
	require.Equal(t, res, &sproto.ResourcesSubscription{})

	res, err = DefaultRMRouter.Allocate(sproto.AllocateRequest{ResourceManager: "bogus"})
	require.Error(t, err)
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
}

func TestValidateResources(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	manager := uuid.NewString()
	retErr := errors.New("error-validate-resources")

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	mockRM.On("ValidateResources", "", 0, true).Return(nil)
	mockRM.On("ValidateResources", "", 1, true).Return(retErr)

	err := ValidateResources(manager, "", 0, true)
	require.NoError(t, err)

	err = ValidateResources(manager, "", 1, true)
	require.Equal(t, retErr, err)

	err = ValidateResources("bogus", "", 0, true)
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
}

func TestDeleteJob(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	manager := uuid.NewString()
	retErr := errors.New("error-delete-job")

	job1 := model.JobID("job1")
	job2 := model.JobID("job2")

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	mockRM.On("DeleteJob", job1).Return(sproto.DeleteJobResponse{}, nil)
	mockRM.On("DeleteJob", job2).Return(sproto.DeleteJobResponse{}, retErr)

	resp, err := DeleteJob(manager, job1)
	require.NoError(t, err)
	require.Equal(t, sproto.DeleteJobResponse{}, resp)

	_, err = DeleteJob(manager, job2)
	require.Equal(t, retErr, err)

	resp, err = DeleteJob("bogus", job1)
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
	require.Equal(t, resp, sproto.EmptyDeleteJobResponse())
}

func TestNotifyContainerRunning(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	manager := uuid.NewString()
	retErr := errors.New("error-container-running")

	req1 := sproto.NotifyContainerRunning{NodeName: "req1"}
	req2 := sproto.NotifyContainerRunning{NodeName: "req2"}

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	mockRM.On("NotifyContainerRunning", req1).Return(nil)
	mockRM.On("NotifyContainerRunning", req2).Return(retErr)

	err := NotifyContainerRunning(manager, req1)
	require.NoError(t, err)

	err = NotifyContainerRunning(manager, req2)
	require.Equal(t, retErr, err)

	err = NotifyContainerRunning("bogus", req1)
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
}

func TestSetGroupWeight(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	manager := uuid.NewString()
	retErr := errors.New("error-set-group-weight")

	req1 := sproto.SetGroupWeight{ResourcePool: "rp1"}
	req2 := sproto.SetGroupWeight{ResourcePool: "rp2"}

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	mockRM.On("SetGroupWeight", req1).Return(nil)
	mockRM.On("SetGroupWeight", req2).Return(retErr)

	err := SetGroupWeight(manager, req1)
	require.NoError(t, err)

	err = SetGroupWeight(manager, req2)
	require.Equal(t, retErr, err)

	err = SetGroupWeight("bogus", req1)
	require.Equal(t, err, ErrRPNotDefined("bogus", "rp1"))
}

func TestSetGroupPriority(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	manager := uuid.NewString()
	retErr := errors.New("error-set-group-weight")

	req1 := sproto.SetGroupPriority{ResourcePool: "rp1"}
	req2 := sproto.SetGroupPriority{ResourcePool: "rp2"}

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	mockRM.On("SetGroupPriority", req1).Return(nil)
	mockRM.On("SetGroupPriority", req2).Return(retErr)

	err := SetGroupPriority(manager, req1)
	require.NoError(t, err)

	err = SetGroupPriority(manager, req2)
	require.Equal(t, retErr, err)

	err = SetGroupPriority("bogus", req1)
	require.Equal(t, err, ErrRPNotDefined("bogus", "rp1"))
}

func TestExternalPreemptionPending(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	manager := uuid.NewString()
	retErr := errors.New("error-ext-preemption")

	alloc1 := model.AllocationID("alloc1")
	alloc2 := model.AllocationID("alloc2")

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	mockRM.On("ExternalPreemptionPending", alloc1).Return(nil)
	mockRM.On("ExternalPreemptionPending", alloc2).Return(retErr)

	err := ExternalPreemptionPending(manager, alloc1)
	require.NoError(t, err)

	err = ExternalPreemptionPending(manager, alloc2)
	require.Equal(t, retErr, err)

	err = ExternalPreemptionPending("bogus", alloc1)
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
}

func TestIsReattachable(t *testing.T) {
	mockRM := mocks.ResourceManager{}
	mockRM.On("IsReattachableOnlyAfterStarted").Return(true)

	manager := uuid.NewString()

	DefaultRMRouter = &MultiRMRouter{rms: map[string]ResourceManager{manager: &mockRM}}

	val := IsReattachableOnlyAfterStarted(manager)
	require.Equal(t, true, val)

	val = IsReattachableOnlyAfterStarted("bogus")
	require.Equal(t, false, val)
}

func TestGetResourcePools(t *testing.T) {
	cases := []struct {
		name     string
		rpNames  []string
		managers int
	}{
		{"simple", []string{uuid.NewString(), uuid.NewString()}, 1},
		{"multirm", []string{uuid.NewString(), uuid.NewString()}, 5},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rms := map[string]ResourceManager{}
			for i := 1; i <= tt.managers; i++ {
				ret := []*resourcepoolv1.ResourcePool{}
				for _, n := range tt.rpNames {
					ret = append(ret, &resourcepoolv1.ResourcePool{Name: n})
				}

				mockRM := mocks.ResourceManager{}
				mockRM.On("GetResourcePools").Return(&apiv1.GetResourcePoolsResponse{ResourcePools: ret}, nil)

				rms[uuid.NewString()] = &mockRM
			}

			DefaultRMRouter = &MultiRMRouter{rms: rms}

			rps, err := DefaultRMRouter.GetResourcePools()
			require.NoError(t, err)
			require.Equal(t, tt.managers*len(tt.rpNames), len(rps.ResourcePools))

			for _, rp := range rps.ResourcePools {
				require.Contains(t, tt.rpNames, rp.Name)
			}
		})
	}
}

func TestGetDefaultComputeResourcePool(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	res := sproto.GetDefaultComputeResourcePoolResponse{PoolName: "default"}

	mockRM.On("GetDefaultComputeResourcePool").Return(res, nil)

	ret, err := GetDefaultComputeResourcePool(manager)
	require.NoError(t, err)
	require.Equal(t, res, ret)

	ret, err = GetDefaultComputeResourcePool("bogus")
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
	require.Empty(t, ret)
}

func TestGetDefaultAuxResourcePool(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	res := sproto.GetDefaultAuxResourcePoolResponse{PoolName: "default"}

	mockRM.On("GetDefaultAuxResourcePool").Return(res, nil)

	ret, err := GetDefaultAuxResourcePool(manager)
	require.NoError(t, err)
	require.Equal(t, res, ret)

	ret, err = GetDefaultAuxResourcePool("bogus")
	require.Equal(t, err, ErrRPNotDefined("bogus", ""))
	require.Empty(t, ret)
}

func TestResolveResourcePool(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rp := "resource-pool"
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	res := sproto.GetDefaultAuxResourcePoolResponse{PoolName: rp}

	mockRM.On("ResolveResourcePool", rp, 1, 1).Return(rp, nil)

	ret, err := ResolveResourcePool(manager, rp, 1, 1)
	require.NoError(t, err)
	require.Equal(t, res.PoolName, ret)

	ret, err = ResolveResourcePool("bogus", rp, 1, 1)
	require.Equal(t, err, ErrRPNotDefined("bogus", rp))
	require.Equal(t, rp, ret)
}

func TestValidateResourcePoolAvailability(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rp := "resource-pool"
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	res := &sproto.ValidateResourcePoolAvailabilityRequest{Name: rp}

	mockRM.On("ValidateResourcePoolAvailability", res).Return([]command.LaunchWarning{}, nil)

	_, err := ValidateResourcePoolAvailability(manager, res)
	require.NoError(t, err)

	_, err = ValidateResourcePoolAvailability("bogus", res)
	require.Equal(t, err, ErrRPNotDefined("bogus", rp))
}

func TestTaskContainerDefaults(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rp := "resource-pool"
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	res := model.TaskContainerDefaultsConfig{}

	mockRM.On("TaskContainerDefaults", rp, res).Return(res, nil)

	_, err := TaskContainerDefaults(manager, rp, res)
	require.NoError(t, err)

	_, err = TaskContainerDefaults("bogus", rp, res)
	require.Equal(t, err, ErrRPNotDefined("bogus", rp))
}

func TestGetJobQ(t *testing.T) {
	rms := map[string]ResourceManager{}
	mockRM := mocks.ResourceManager{}

	manager := uuid.NewString()
	rp := "resource-pool"
	rms[manager] = &mockRM
	DefaultRMRouter = &MultiRMRouter{rms: rms}

	res := map[model.JobID]*sproto.RMJobInfo{}

	mockRM.On("GetJobQ", rp).Return(res, nil)

	ret, err := GetJobQ(manager, rp)
	require.NoError(t, err)
	require.Equal(t, ret, res)

	ret, err = GetJobQ("bogus", rp)
	require.Equal(t, err, ErrRPNotDefined("bogus", rp))
	require.Nil(t, ret)
}

func mockRMConfig(pools int) config.ResourceManagerWithPoolsConfig {
	uuid := uuid.NewString()
	return config.ResourceManagerWithPoolsConfig{
		ResourceManager: &config.ResourceManagerConfig{
			AgentRM: &config.AgentResourceManagerConfig{
				Scheduler: &config.SchedulerConfig{
					FairShare:     &config.FairShareSchedulerConfig{},
					FittingPolicy: "best",
				},
				DefaultAuxResourcePool:     fmt.Sprintf("default-aux-%s", uuid),
				DefaultComputeResourcePool: fmt.Sprintf("default-compute-%s", uuid),
				Name:                       uuid,
			},
		},
		ResourcePools: mockResourcePools(pools),
	}
}

func mockResourcePools(pools int) []config.ResourcePoolConfig {
	var res []config.ResourcePoolConfig
	uuid := uuid.NewString()
	for i := 1; i <= pools; i++ {
		res = append(res, config.ResourcePoolConfig{
			PoolName: uuid,
		})
	}
	return res
}
