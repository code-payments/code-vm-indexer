package geyser

import (
	"time"

	"github.com/code-payments/code-server/pkg/config"
	"github.com/code-payments/code-server/pkg/config/env"
)

const (
	envConfigPrefix = "GEYSER_WORKER_"

	GrpcPluginEndointConfigEnvName = envConfigPrefix + "GRPC_PLUGIN_ENDPOINT"
	defaultGrpcPluginEndoint       = ""

	ProgramUpdateWorkerCountConfigEnvName = envConfigPrefix + "PROGRAM_UPDATE_WORKER_COUNT"
	defaultProgramUpdateWorkerCount       = 1024

	ProgramUpdateQueueSizeConfigEnvName = envConfigPrefix + "PROGRAM_UPDATE_QUEUE_SIZE"
	defaultProgramUpdateQueueSize       = 1_000_000

	VmAccountConfigEnvName = envConfigPrefix + "VM_ACCOUNT"
	defaultVmAccount       = ""

	MemoryAccountBackupWorkerIntervalConfigEnvName = envConfigPrefix + "MEMORY_ACCOUNT_BACKUP_WORKER_INTERVAL"
	defaultMemoryAccountBackupWorkerInterval       = time.Minute
)

type conf struct {
	grpcPluginEndpoint config.String

	programUpdateWorkerCount config.Uint64
	programUpdateQueueSize   config.Uint64

	vmAccount config.String

	memoryAccountBackkupWorkerInterval config.Duration
}

// ConfigProvider defines how config values are pulled
type ConfigProvider func() *conf

// WithEnvConfigs returns configuration pulled from environment variables
func WithEnvConfigs() ConfigProvider {
	return func() *conf {
		return &conf{
			grpcPluginEndpoint: env.NewStringConfig(GrpcPluginEndointConfigEnvName, defaultGrpcPluginEndoint),

			programUpdateWorkerCount: env.NewUint64Config(ProgramUpdateWorkerCountConfigEnvName, defaultProgramUpdateWorkerCount),
			programUpdateQueueSize:   env.NewUint64Config(ProgramUpdateQueueSizeConfigEnvName, defaultProgramUpdateQueueSize),

			vmAccount: env.NewStringConfig(VmAccountConfigEnvName, defaultVmAccount),

			memoryAccountBackkupWorkerInterval: env.NewDurationConfig(MemoryAccountBackupWorkerIntervalConfigEnvName, defaultMemoryAccountBackupWorkerInterval),
		}
	}
}
