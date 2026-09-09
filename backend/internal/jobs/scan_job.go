package jobs

import (
	"context"
	"errors"
	"fmt"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"strings"
	"time"
)

type ScanRunner interface {
	RunScan(context.Context, domain.Scan, string) error
}
type ScanJob struct {
	Queue  Queue
	Runner ScanRunner
}

func (j ScanJob) ProcessOne(ctx context.Context) (bool, error) {
	scan, url, err := j.Queue.Claim(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim scan: %w", err)
	}
	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err = j.Runner.RunScan(runCtx, scan, url); err != nil {
		// A cancellation must not prevent the terminal failure being recorded.
		failCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer stop()
		if failure := j.Queue.Fail(failCtx, scan.ID, errorCode(err), err.Error()); failure != nil {
			return true, fmt.Errorf("scan %s: %v; persist failure: %w", scan.ID, err, failure)
		}
	}
	return true, nil
}
func errorCode(err error) string {
	for _, code := range []string{"UNSUPPORTED_INPUT", "INVALID_REPOSITORY_URL", "GITHUB_UNAVAILABLE", "GITHUB_METADATA_ERROR", "GITHUB_DOWNLOAD_ERROR", "REPOSITORY_FILE_TOO_LARGE"} {
		if strings.Contains(err.Error(), code) {
			return code
		}
	}
	return "SCAN_FAILED"
}
