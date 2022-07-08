// Package tracing enables [Datadog APM tracing](https://docs.datadoghq.com/tracing/) capabilities,
// focusing specifically on [Error Tracking](https://docs.datadoghq.com/tracing/error_tracking/)
package tracing

import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

var (
	// reimport relevant bits of the tracer API
	StartSpan = tracer.StartSpan
	WithError = tracer.WithError
)

// Start boots the datadog tracer, run this once early in the startup sequence.
func Start() {
	tracer.Start(tracer.WithService("xmtp-node"))
}

// Stop shuts down the datadog tracer, defer this right after Start().
func Stop() {
	tracer.Stop()
}

// Do executes action in the context of a top level span,
// tagging the span with the error if the action panics.
// This should trigger DD APM's Error Tracking to record the error.
func Do(spanName string, action func()) {
	span := StartSpan(spanName)
	defer func() {
		r := recover()
		switch r := r.(type) {
		case error:
			// If action panics with an error,
			// finish the span with the error.
			span.Finish(WithError(r))
		default:
			// This is the normal non-panicking path
			// as well as path with panics that don't have an error.
			span.Finish()
		}
		if r != nil {
			// Repanic so that we don't suppress normal panic behavior.
			panic(r)
		}
	}()
	action()
}
