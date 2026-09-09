package jobs

import (
	"context"
	"log"
	"time"
)

type Worker struct{ Job ScanJob }

func (w Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		worked, err := w.Job.ProcessOne(ctx)
		if err != nil {
			log.Printf("scan worker: %v", err)
		}
		if !worked || err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}
}
