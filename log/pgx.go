package log

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"time"
)

type sqlTracer struct {
	logger zerolog.Logger
}

type sqlTraceData struct {
	pgx.TraceQueryStartData
	startTime time.Time
}

type dbTraceContextKey struct{}

func NewSQLTracer(logger zerolog.Logger) pgx.QueryTracer {
	return &sqlTracer{logger: logger}
}

func (t *sqlTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(
		ctx,
		dbTraceContextKey{},
		&sqlTraceData{
			TraceQueryStartData: data,
			startTime:           time.Now(),
		},
	)
}

func (t *sqlTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	trace := ctx.Value(dbTraceContextKey{}).(*sqlTraceData)
	t.logger.Debug().
		Str("status", data.CommandTag.String()).
		Str("query", trace.SQL).
		Any("args", trace.Args).
		Msg("")
}
