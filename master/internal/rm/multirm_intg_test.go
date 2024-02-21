//go:build integration
// +build integration

package rm

import (
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
	"github.com/determined-ai/determined/master/pkg/etc"
	"github.com/determined-ai/determined/master/pkg/model"
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
		{"multirm", []string{uuid.NewString(), uuid.NewString()}, 5},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rms := map[string]ResourceManager{}
			for i := 1; i <= tt.managers; i++ {
				ret := map[model.AllocationID]sproto.AllocationSummary{}
				for _, alloc := range tt.allocNames {
					a := alloc
					ret[*model.NewAllocationID(&a)] = sproto.AllocationSummary{}
				}
				require.Equal(t, len(ret), len(tt.allocNames))

				mockRM := mocks.ResourceManager{}
				mockRM.On("GetAllocationSummaries").Return(ret, nil)

				rms[uuid.NewString()] = &mockRM
			}

			DefaultRMRouter = &MultiRMRouter{rms: rms}

			allocs, err := GetAllocationSummaries()
			require.NoError(t, err)
			require.Equal(t, tt.managers*len(tt.allocNames), len(allocs))
			require.NotNil(t, allocs)

			bogus := "bogus"
			require.Empty(t, allocs[*model.NewAllocationID(&bogus)])

			for _, name := range tt.allocNames {
				n := name
				require.NotNil(t, allocs[*model.NewAllocationID(&n)])
			}
		})
	}
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
