package geyser

import (
	"context"
	"time"

	"github.com/code-payments/ocp-server/retry"
	"github.com/code-payments/ocp-server/retry/backoff"
)

var waitForFinalizationRetryStrategies = []retry.Strategy{
	retry.NonRetriableErrors(context.Canceled),
	retry.Limit(30),
	retry.Backoff(backoff.Constant(3*time.Second), 3*time.Second),
}
