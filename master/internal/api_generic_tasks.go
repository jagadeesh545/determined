package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ghodss/yaml"

	"github.com/determined-ai/determined/master/internal/api"
	"github.com/determined-ai/determined/master/internal/authz"
	"github.com/determined-ai/determined/master/internal/command"
	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/internal/grpcutil"
	"github.com/determined-ai/determined/master/internal/job/jobservice"
	"github.com/determined-ai/determined/master/internal/project"
	"github.com/determined-ai/determined/master/internal/rm/tasklist"
	"github.com/determined-ai/determined/master/internal/sproto"
	"github.com/determined-ai/determined/master/internal/task"
	"github.com/determined-ai/determined/master/internal/user"
	"github.com/determined-ai/determined/master/pkg/check"
	pkgCommand "github.com/determined-ai/determined/master/pkg/command"
	"github.com/determined-ai/determined/master/pkg/logger"
	"github.com/determined-ai/determined/master/pkg/model"
	"github.com/determined-ai/determined/master/pkg/ptrs"
	"github.com/determined-ai/determined/master/pkg/tasks"
	"github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/projectv1"
	"github.com/determined-ai/determined/proto/pkg/utilv1"
)

func (a *apiServer) getGenericTaskLaunchParameters(
	ctx context.Context,
	contextDirectory []*utilv1.File,
	configYAML string,
	projectID int,
	parentConfig []byte,
) (
	*tasks.GenericTaskSpec, []pkgCommand.LaunchWarning, []byte, error,
) {
	genericTaskSpec := &tasks.GenericTaskSpec{
		ProjectID: projectID,
	}

	// Validate the userModel and get the agent userModel group.
	userModel, _, err := grpcutil.GetUser(ctx)
	if err != nil {
		return nil,
			nil,
			nil,
			status.Errorf(codes.Unauthenticated, "failed to get the user: %s", err)
	}

	proj, err := a.GetProjectByID(ctx, int32(genericTaskSpec.ProjectID), *userModel)
	if err != nil {
		return nil, nil, nil, err
	}
	agentUserGroup, err := user.GetAgentUserGroup(ctx, userModel.ID, int(proj.WorkspaceId))
	if err != nil {
		return nil, nil, nil, err
	}

	// Validate the resource configuration.
	resources := model.ParseJustResources([]byte(configYAML))

	if resources.SlotsPerTask == nil {
		resources.SlotsPerTask = ptrs.Ptr(1)
	}

	poolName, launchWarnings, err := ResolveResources(a, resources.ResourcePool, *resources.SlotsPerTask, int(proj.WorkspaceId))
	if err != nil {
		return nil, nil, nil, err
	}
	// Get the base TaskSpec.
	taskSpec, err := fillTaskSpec(a, poolName, agentUserGroup, userModel)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get the full configuration.
	taskConfig := model.DefaultConfigGenericTaskConfig(&taskSpec.TaskContainerDefaults)
	workDirInDefaults := taskConfig.WorkDir

	if len(parentConfig) != 0 {
		if err := yaml.Unmarshal(parentConfig, &taskConfig); err != nil {
			return nil, nil, nil, fmt.Errorf("yaml unmarshaling generic task config: %w", err)
		}
	}

	if err := yaml.Unmarshal([]byte(configYAML), &taskConfig); err != nil {
		return nil, nil, nil, fmt.Errorf("yaml unmarshaling generic task config: %w", err)
	}

	// Copy discovered (default) resource pool name and slot count.

	fillTaskConfig(&taskConfig.Resources.RawResourcePool, poolName, &taskConfig.Resources.RawSlotsPerTask, *resources.SlotsPerTask, taskSpec, &taskConfig.Environment)

	var contextDirectoryBytes []byte
	contextDirectoryBytes, err = fillContextDir(&taskConfig.WorkDir, workDirInDefaults, contextDirectory)
	if err != nil {
		return nil, nil, nil, err
	}

	var token string
	token, err = getTaskSessionToken(ctx, userModel)
	if err != nil {
		return nil, nil, nil, err
	}

	taskSpec.UserSessionToken = token

	genericTaskSpec.Base = taskSpec
	genericTaskSpec.GenericTaskConfig = taskConfig

	genericTaskSpec.Base.ExtraEnvVars = map[string]string{
		"DET_TASK_TYPE": string(model.TaskTypeGeneric),
	}

	return genericTaskSpec, launchWarnings, contextDirectoryBytes, nil
}

func (a *apiServer) canCreateGenericTask(ctx context.Context, projectID int) error {
	userModel, _, err := grpcutil.GetUser(ctx)
	if err != nil {
		return err
	}

	errProjectNotFound := api.NotFoundErrs("project", fmt.Sprint(projectID), true)
	p := &projectv1.Project{}
	if err := a.m.db.QueryProto("get_project", p, projectID); errors.Is(err, db.ErrNotFound) {
		return errProjectNotFound
	} else if err != nil {
		return err
	}
	if err := project.AuthZProvider.Get().CanGetProject(ctx, *userModel, p); err != nil {
		return authz.SubIfUnauthorized(err, errProjectNotFound)
	}

	if err := command.AuthZProvider.Get().CanCreateGenericTask(
		ctx, *userModel, model.AccessScopeID(p.WorkspaceId)); err != nil {
		return status.Errorf(codes.PermissionDenied, err.Error())
	}

	return nil
}

func (a *apiServer) CreateGenericTask(
	ctx context.Context, req *apiv1.CreateGenericTaskRequest,
) (*apiv1.CreateGenericTaskResponse, error) {
	var projectID int
	if req.ProjectId != nil {
		projectID = int(*req.ProjectId)
	} else {
		projectID = model.DefaultProjectID
	}

	if err := a.canCreateGenericTask(ctx, projectID); err != nil {
		return nil, err
	}

	var parentConfig []byte
	if req.ForkedFromId != nil {
		// Can't use getExperimentAndCheckDoActions since model.Experiment doesn't have ParentArchived.
		getTaskReq := &apiv1.GetTaskRequest{
			TaskId: *req.ForkedFromId,
		}
		resp, err := a.GetTask(ctx, getTaskReq)
		if err != nil {
			return nil, err
		}

		parentConfig, err = resp.Task.Config.MarshalJSON()
		if err != nil {
			return nil, err
		}
	}

	genericTaskSpec, warnings, contextDirectoryBytes, err := a.getGenericTaskLaunchParameters(
		ctx, req.ContextDirectory, req.Config, projectID, parentConfig,
	)
	if err != nil {
		return nil, err
	}

	if err := check.Validate(genericTaskSpec.GenericTaskConfig); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid generic task config: %s",
			err.Error(),
		)
	}

	// Persist the task.
	taskID := model.NewTaskID()
	jobID := model.NewJobID()
	startTime := time.Now()
	err = db.Bun().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := db.AddJobTx(ctx, tx, &model.Job{
			JobID:   jobID,
			JobType: model.JobTypeGeneric,
			OwnerID: &genericTaskSpec.Base.Owner.ID,
		}); err != nil {
			return fmt.Errorf("persisting job %v: %w", taskID, err)
		}

		if err := db.AddTaskTx(ctx, tx, &model.Task{
			TaskID:     taskID,
			TaskType:   model.TaskTypeGeneric,
			StartTime:  startTime,
			JobID:      &jobID,
			LogVersion: model.CurrentTaskLogVersion,
		}, &genericTaskSpec.GenericTaskConfig); err != nil {
			return fmt.Errorf("persisting task %v: %w", taskID, err)
		}

		// TODO persist config elemnts

		if _, err := tx.NewInsert().Model(&model.TaskContextDirectory{
			TaskID:           taskID,
			ContextDirectory: contextDirectoryBytes,
		}).Exec(ctx); err != nil {
			return fmt.Errorf("persisting context directory files: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("persisting task information: %w", err)
	}

	logCtx := logger.Context{
		"job-id":    jobID,
		"task-id":   taskID,
		"task-type": model.TaskTypeGeneric,
	}
	priorityChange := func(priority int) error {
		return nil
	}
	if err = tasklist.GroupPriorityChangeRegistry.Add(jobID, priorityChange); err != nil {
		return nil, err
	}

	err = task.DefaultService.StartAllocation(logCtx, sproto.AllocateRequest{
		AllocationID:      model.AllocationID(fmt.Sprintf("%s.%d", taskID, 1)),
		TaskID:            taskID,
		JobID:             jobID,
		JobSubmissionTime: startTime,
		IsUserVisible:     true,
		Name:              fmt.Sprintf("Generic Task %s", taskID),

		SlotsNeeded:  genericTaskSpec.GenericTaskConfig.Resources.SlotsPerTask(),
		ResourcePool: genericTaskSpec.GenericTaskConfig.Resources.ResourcePool(),
		FittingRequirements: sproto.FittingRequirements{
			SingleAgent: genericTaskSpec.GenericTaskConfig.Resources.IsSingleNode(),
		},

		Restore: false,
	}, a.m.db, a.m.rm, genericTaskSpec, onAllocationExit)
	if err != nil {
		return nil, err
	}

	jobservice.DefaultService.RegisterJob(jobID, genericTaskSpec)

	return &apiv1.CreateGenericTaskResponse{
		TaskId:   string(taskID),
		Warnings: pkgCommand.LaunchWarningToProto(warnings),
	}, nil
}

func onAllocationExit(ae *task.AllocationExited) {}
