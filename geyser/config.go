package geyser

import (
	"context"
	"strings"
	"time"

	"github.com/code-payments/ocp-server/config"
	"github.com/code-payments/ocp-server/config/env"
)

const (
	envConfigPrefix = "GEYSER_WORKER_"

	GrpcPluginEndointConfigEnvName = envConfigPrefix + "GRPC_PLUGIN_ENDPOINT"
	defaultGrpcPluginEndoint       = ""

	GrpcPluginXTokenConfigEnvName = envConfigPrefix + "GRPC_PLUGIN_X_TOKEN"
	defaultGrpcPluginXToken       = ""

	ProgramUpdateWorkerCountConfigEnvName = envConfigPrefix + "PROGRAM_UPDATE_WORKER_COUNT"
	defaultProgramUpdateWorkerCount       = 1024

	ProgramUpdateQueueSizeConfigEnvName = envConfigPrefix + "PROGRAM_UPDATE_QUEUE_SIZE"
	defaultProgramUpdateQueueSize       = 1_000_000

	VmAccountsConfigEnvName = envConfigPrefix + "VM_ACCOUNTS"
	defaultVmAccounts       = ""

	MemoryAccountBackupWorkerIntervalConfigEnvName = envConfigPrefix + "MEMORY_ACCOUNT_BACKUP_WORKER_INTERVAL"
	defaultMemoryAccountBackupWorkerInterval       = time.Minute
)

type conf struct {
	grpcPluginEndpoint config.String
	grpcPluginXToken   config.String

	programUpdateWorkerCount config.Uint64
	programUpdateQueueSize   config.Uint64

	vmAccounts config.String

	memoryAccountBackkupWorkerInterval config.Duration
}

// ConfigProvider defines how config values are pulled
type ConfigProvider func() *conf

// WithEnvConfigs returns configuration pulled from environment variables
func WithEnvConfigs() ConfigProvider {
	return func() *conf {
		return &conf{
			grpcPluginEndpoint: env.NewStringConfig(GrpcPluginEndointConfigEnvName, defaultGrpcPluginEndoint),
			grpcPluginXToken:   env.NewStringConfig(GrpcPluginXTokenConfigEnvName, defaultGrpcPluginXToken),

			programUpdateWorkerCount: env.NewUint64Config(ProgramUpdateWorkerCountConfigEnvName, defaultProgramUpdateWorkerCount),
			programUpdateQueueSize:   env.NewUint64Config(ProgramUpdateQueueSizeConfigEnvName, defaultProgramUpdateQueueSize),

			vmAccounts: env.NewStringConfig(VmAccountsConfigEnvName, defaultVmAccounts),

			memoryAccountBackkupWorkerInterval: env.NewDurationConfig(MemoryAccountBackupWorkerIntervalConfigEnvName, defaultMemoryAccountBackupWorkerInterval),
		}
	}
}

func parseVmAccountsConfig(ctx context.Context, c *conf) []string {
	parts := strings.Split(c.vmAccounts.Get(ctx), ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}
