package tasks

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/determined-ai/determined/cluster/pkg/archive"
	"github.com/determined-ai/determined/cluster/pkg/cproto"
	"github.com/determined-ai/determined/cluster/pkg/etc"
	"github.com/determined-ai/determined/cluster/pkg/model"
	"github.com/determined-ai/determined/cluster/pkg/schemas"
	"github.com/determined-ai/determined/cluster/pkg/schemas/expconf"
	"github.com/determined-ai/determined/cluster/pkg/ssh"
)

// genericCommandSpecMetadata is GenericCommandSpec.Metadata.
type genericCommandSpecMetadata struct {
	PrivateKey    *string             `json:"privateKey"`
	PublicKey     *string             `json:"publicKey"`
	ExperimentIDs []int32             `json:"experiment_ids"`
	TrialIDs      []int32             `json:"trial_ids"`
	WorkspaceID   model.AccessScopeID `json:"workspace_id"`
}

// MarshalToMap converts typed struct into a map.
func (metadata *genericCommandSpecMetadata) MarshalToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GenericCommandSpec is a description of a task for running a command.
type GenericCommandSpec struct {
	Base TaskSpec

	CommandID string

	Config model.CommandConfig
	// Deprecated: kept so we can still marshal to this.
	// Please use command.CreateGeneric.modelDef instead.
	UserFiles       archive.Archive
	AdditionalFiles archive.Archive
	Metadata        genericCommandSpecMetadata

	Keys *ssh.PrivateAndPublicKeys

	WatchProxyIdleTimeout  bool
	WatchRunnerIdleTimeout bool

	TaskType model.TaskType
}

// ToTaskSpec generates a TaskSpec.
func (s GenericCommandSpec) ToTaskSpec() TaskSpec {
	res := s.Base

	s.MakeEnvPorts()
	res.Environment = s.Config.Environment.ToExpconf()

	res.ResourcesConfig = s.Config.Resources.ToExpconf()

	res.PbsConfig = s.Config.Pbs

	res.SlurmConfig = s.Config.Slurm

	res.WorkDir = DefaultWorkDir
	if s.Config.WorkDir != nil {
		res.WorkDir = *s.Config.WorkDir
	}
	res.ResolveWorkDir()

	if s.Keys != nil {
		s.AdditionalFiles = append(s.AdditionalFiles, archive.Archive{
			res.AgentUserGroup.OwnedArchiveItem(sshDir, nil, sshDirMode, tar.TypeDir),
			res.AgentUserGroup.OwnedArchiveItem(
				shellAuthorizedKeysFile, s.Keys.PublicKey, 0o644, tar.TypeReg,
			),
			res.AgentUserGroup.OwnedArchiveItem(
				privKeyFile, s.Keys.PrivateKey, privKeyMode, tar.TypeReg,
			),
			res.AgentUserGroup.OwnedArchiveItem(
				pubKeyFile, s.Keys.PublicKey, pubKeyMode, tar.TypeReg,
			),
			res.AgentUserGroup.OwnedArchiveItem(
				sshdConfigFile,
				etc.MustStaticFile(etc.SSHDConfigResource),
				0o644,
				tar.TypeReg,
			),
		}...)
	}

	res.ExtraArchives = []cproto.RunArchive{
		wrapArchive(s.AdditionalFiles, rootDir),
	}

	res.Description = fmt.Sprintf("cmd-%s", s.CommandID)

	res.Entrypoint = s.Config.Entrypoint

	res.Mounts = ToDockerMounts(s.Config.BindMounts.ToExpconf(), res.WorkDir)

	if shm := s.Config.Resources.ShmSize; shm != nil {
		res.ShmSize = int64(*shm)
	}

	res.TaskType = s.TaskType

	// Evict the context from memory after starting the command as it is no longer needed. We
	// evict as soon as possible to prevent the master from hitting an OOM.
	s.AdditionalFiles = nil

	return res
}

// TrialSpec has matching `ProxyPorts` and `MakeEnvPorts` methods. Long-term, we should
// - unify TrialSpec and GenericCommandSpec
// - move CommandConfig to expconf

// MakeEnvPorts fills in `Environment.Ports` i.e. exposed ports for container config.
func (s *GenericCommandSpec) MakeEnvPorts() {
	if s.Config.Environment.Ports != nil {
		panic("CommandSpec Environment.Ports are only supposed to be generated once.")
	}

	ppc := s.ProxyPorts()
	s.Config.Environment.Ports = map[string]int{}
	for _, pp := range ppc {
		port := pp.ProxyPort()
		s.Config.Environment.Ports[strconv.Itoa(port)] = port
	}
}

// ProxyPorts combines user-defined and system proxy configs.
func (s *GenericCommandSpec) ProxyPorts() expconf.ProxyPortsConfig {
	env := schemas.WithDefaults(s.Config.Environment.ToExpconf())
	epp := schemas.WithDefaults(s.Base.ExtraProxyPorts)
	out := make(expconf.ProxyPortsConfig, 0, len(epp)+len(env.ProxyPorts()))

	for _, pp := range epp {
		out = append(out, pp)
	}

	for _, pp := range env.ProxyPorts() {
		out = append(out, pp)
	}

	return out
}
