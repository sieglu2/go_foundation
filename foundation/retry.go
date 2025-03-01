package foundation

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
)

func DoRetry(
	ctx context.Context, execTimeout time.Duration,
	maxTimes int, backoffWait time.Duration,
	exec func(timeoutCtx context.Context) error,
) error {
	logger := Logger()

	backoffConfig := backoff.WithMaxRetries(backoff.NewConstantBackOff(backoffWait), uint64(maxTimes))

	err := backoff.Retry(func() error {
		// Change this part
		if ctx.Err() != nil {
			return backoff.Permanent(ctx.Err()) // Mark as permanent error to stop retrying
		}

		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, execTimeout)
		defer timeoutCancel() // Better to use defer here

		insideErr := exec(timeoutCtx)
		if insideErr != nil {
			return insideErr // Let backoff decide if it should retry
		}

		return nil
	}, backoffConfig)

	if err != nil {
		err = fmt.Errorf("failed retry eventually: %v", err)
		logger.Error(err)
		return err
	}

	return nil
}
