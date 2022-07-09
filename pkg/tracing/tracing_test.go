package tracing

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GoPanicWrap_WaitGroup(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	finished := false
	GoPanicWrap(ctx, &wg, "test", func(ctx context.Context) {
		<-ctx.Done()
		finished = true
	})
	done := false
	go func() {
		wg.Wait()
		done = true
	}()
	go func() { time.Sleep(time.Millisecond); cancel() }()

	assert.Eventually(t, func() bool { return finished && done }, time.Second, 10*time.Millisecond)
}
