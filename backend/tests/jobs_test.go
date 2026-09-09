package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/jobs"
)

type failedRunner struct {
	err    error
	cancel context.CancelFunc
}

func (r failedRunner) RunScan(context.Context, domain.Scan, string) error {
	if r.cancel != nil {
		r.cancel()
	}
	return r.err
}

type testQueue struct {
	claimErr, failErr error
	code, message     string
	failureContextErr error
}

func (q *testQueue) Claim(context.Context) (domain.Scan, string, error) {
	return domain.Scan{ID: "test"}, "https://github.com/controlled/test", q.claimErr
}
func (q *testQueue) Fail(ctx context.Context, id, code, message string) error {
	q.code = code
	q.message = message
	q.failureContextErr = ctx.Err()
	return q.failErr
}
func TestWorkerFailureHandling(t *testing.T) {
	for _, code := range []string{"GITHUB_UNAVAILABLE", "GITHUB_DOWNLOAD_ERROR", "SCAN_FAILED"} {
		t.Run(code, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			queue := &testQueue{}
			job := jobs.ScanJob{Queue: queue, Runner: failedRunner{err: errors.New(code + ": injected failure"), cancel: cancel}}
			worked, err := job.ProcessOne(ctx)
			if !worked || err != nil || queue.code != code || queue.failureContextErr != nil {
				t.Fatalf("failure not recorded after cancellation: %t %v %+v", worked, err, queue)
			}
		})
	}
	queue := &testQueue{claimErr: errors.New("database disconnected")}
	if worked, err := (jobs.ScanJob{Queue: queue}).ProcessOne(context.Background()); worked || err == nil {
		t.Fatal("claim error swallowed")
	}
	queue = &testQueue{failErr: errors.New("database disconnected")}
	job := jobs.ScanJob{Queue: queue, Runner: failedRunner{err: errors.New("graph unavailable")}}
	if _, err := job.ProcessOne(context.Background()); err == nil || !strings.Contains(err.Error(), "persist failure") {
		t.Fatal("failure persistence error swallowed")
	}
}
