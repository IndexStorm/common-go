package telemetry

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
)

var (
	_ pgx.QueryTracer   = (*SqlTracer)(nil)
	_ pgx.ConnectTracer = (*SqlTracer)(nil)
)

type SqlTracer struct {
	tracer trace.Tracer
	attrs  []attribute.KeyValue
}

func NewPgxTracer(attrs ...attribute.KeyValue) *SqlTracer {
	// DBOperationBatchSizeKey
	return &SqlTracer{
		tracer: otel.Tracer("github.com/IndexStorm/common-go/pgx",
			trace.WithInstrumentationVersion("v1.0.1"),
			trace.WithSchemaURL(semconv.SchemaURL),
		),
		attrs: attrs,
	}
}

func (t *SqlTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}
	attrs := make([]attribute.KeyValue, 0, 3+len(t.attrs)+len(data.Args))
	attrs = append(attrs, semconv.DBQueryText(data.SQL))
	fields := strings.Fields(data.SQL)
	var summary string
	if len(fields) < 2 {
		summary = data.SQL
	} else {
		const INSERT = "INSERT"
		const SELECT = "SELECT"
		const FROM = "FROM"
		var collectionName string
		if strings.ToUpper(fields[0]) == INSERT {
			collectionName = fields[2]
			summary = `INSERT INTO ` + fields[2]
		} else if strings.ToUpper(fields[0]) == SELECT {
			fromFound := false
			for _, field := range fields {
				if strings.ToUpper(field) == FROM {
					fromFound = true
					continue
				} else if fromFound {
					collectionName = field
					break
				}
			}
			summary = `SELECT FROM ` + collectionName
		} else {
			collectionName = fields[1]
			summary = fields[0] + " " + fields[1]
		}
		attrs = append(attrs,
			semconv.DBOperationName(fields[0]),
			semconv.DBCollectionName(collectionName),
			attribute.String("db.query.summary", summary),
		)
	}
	for i, arg := range data.Args {
		key := "db.operation.parameter." + strconv.Itoa(i)
		attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", arg)))
	}
	attrs = append(attrs, t.attrs...)
	ctx, _ = t.tracer.Start(ctx, summary, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
	return ctx
}

func (t *SqlTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	err := data.Err
	t.recordQueryError(span, err)
	t.recordRows(span, err, data.CommandTag.RowsAffected())
	span.End()
}

func (t *SqlTracer) TraceConnectStart(ctx context.Context, _ pgx.TraceConnectStartData) context.Context {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}
	attrs := make([]attribute.KeyValue, 1+len(t.attrs))
	attrs = append(attrs, semconv.DBOperationName("connect"))
	attrs = append(attrs, t.attrs...)
	ctx, _ = t.tracer.Start(ctx, "connect", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
	return ctx
}

func (t *SqlTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	span := trace.SpanFromContext(ctx)
	t.recordError(span, data.Err)
	span.End()
}

func (t *SqlTracer) recordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.SetAttributes(attribute.String("db.response.error.code", pgErr.Code))
		}
		span.SetAttributes(attribute.String("db.response.error.summary", err.Error()))
	}
}

func (t *SqlTracer) recordQueryError(span trace.Span, err error) {
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			span.SetAttributes(attribute.String("db.response.error.code", pgErr.Code))
		}
		span.SetAttributes(attribute.String("db.response.error.summary", err.Error()))
	}
}

func (t *SqlTracer) recordRows(span trace.Span, err error, rows int64) {
	if err == nil {
		span.SetAttributes(attribute.Int64("db.response.returned_rows", rows))
	} else if errors.Is(err, pgx.ErrNoRows) {
		span.SetAttributes(attribute.Int64("db.response.returned_rows", 0))
	}
}
