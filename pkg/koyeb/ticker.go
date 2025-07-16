package koyeb

import (
	"context"
	"iter"
	"time"
)

func ticker(ctx context.Context, interval time.Duration) iter.Seq[time.Time] {
	t := time.NewTicker(interval)
	return func(yield func(time.Time) bool) {
		defer t.Stop()
		for {
			select {
			case e := <-t.C:
				if !yield(e) {
					return
				}
			case <-ctx.Done():
				return
			}

		}
	}

}
