// Package tracing enables [Datadog APM tracing](https://docs.datadoghq.com/tracing/) capabilities,
// focusing specifically on [Error Tracking](https://docs.datadoghq.com/tracing/error_tracking/)
package tracing

import (
	"context"
	"sync"

	"github.com/xmtp/xmtp-node-go/pkg/logging"
	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// reimport relevant bits of the tracer API
var (
	StartSpanFromContext = tracer.StartSpanFromContext
	StartSpan            = tracer.StartSpan
	ChildOf              = tracer.ChildOf
	SpanType             = tracer.SpanType
	WithError            = tracer.WithError
	ContextWithSpan      = tracer.ContextWithSpan
)

type Span = tracer.Span

type logger struct{ *zap.Logger }

func (l logger) Log(msg string) {
	l.Error(msg)
}

// Start boots the datadog tracer, run this once early in the startup sequence.
func Start(version string, l *zap.Logger) {
	tracer.Start(
		tracer.WithService("xmtp-node"),
		tracer.WithServiceVersion(version),
		tracer.WithLogger(logger{l}),
		tracer.WithRuntimeMetrics(),
	)
}

// Stop shuts down the datadog tracer, defer this right after Start().
func Stop() {
	tracer.Stop()
}

// Do executes action in the context of a span.
// Tags the span with the error if action returns one.
func Do(ctx context.Context, spanName string, action func(context.Context, Span) error) error {
	span, ctx := tracer.StartSpanFromContext(ctx, spanName)
	defer span.Finish()
	log := logging.From(ctx).With(zap.String("span", spanName))
	ctx = logging.With(ctx, Link(span, log))
	err := action(ctx, span)
	if err != nil {
		span.Finish(WithError(err))
	}
	return err
}

// PanicDo executes the body guarding for panics.
// If panic happens it emits a span with the error attached.
// This should trigger DD APM's Error Tracking to record the error.
func PanicsDo(ctx context.Context, name string, body func(context.Context)) {
	defer func() {
		r := recover()
		if err, ok := r.(error); ok {
			StartSpan("panic: " + name).Finish(
				WithError(err),
			)
		}
		if r != nil {
			// Repanic so that we don't suppress normal panic behavior.
			panic(r)
		}
	}()
	body(ctx)
}

// Link connects a logger to a particular trace and span.
// DD APM should provide some additional functionality based on that.
func Link(span tracer.Span, l *zap.Logger) *zap.Logger {
	return l.With(
		zap.Uint64("dd.trace_id", span.Context().TraceID()),
		zap.Uint64("dd.span_id", span.Context().SpanID()))
}

// Run the action in a goroutine, synchronize the goroutine exit with the WaitGroup,
// The action must respect cancellation of the Context.
func GoPanicsDo(ctx context.Context, wg *sync.WaitGroup, name string, action func(context.Context)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		PanicsDo(ctx, name, action)
	}()
}
