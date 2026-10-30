package logging

/*
	This is the custom logging module that will be used throughout the project
	for event logging and context tracing. It is meant to provide an interface
	such that it can be changed in the future to allow for different loggers
	(i.e. Datadog)
*/

import (
	"context"
	"log/slog"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
)

type Context context.Context

func StartSpan(ctx Context, spanName string) (context.Context, trace.Span) {
	return tracer.Start(
		ctx,
		spanName,
	)
}

func Debug(ctx Context, msg string) {
	slog.DebugContext(
		ctx,
		msg,
	)
}

func Info(ctx Context, msg string) {
	slog.InfoContext(
		ctx,
		msg,
	)
}

func Warn(ctx Context, msg string) {
	slog.WarnContext(
		ctx,
		msg,
	)
}

func Error(ctx Context, msg string) {
	slog.ErrorContext(
		ctx,
		msg,
	)
}

func init() {
	// Establish initial logging context
	ctx := context.Background()

	uptrace.ConfigureOpentelemetry(
		uptrace.WithServiceName("COBI"),
	)
	defer uptrace.Shutdown(ctx)

	slog.SetDefault(otelslog.NewLogger("COBI"))

	tracer = otel.Tracer("COBI")
}
