package telemetry

import (
	"context"
	"encoding/binary"
	"go.opentelemetry.io/otel/trace"
	"hash/maphash"
)

type lockFreeIdGenerator struct {
}

func (g *lockFreeIdGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	data := make([]byte, 0, 24)
	data = binary.BigEndian.AppendUint64(data[:], new(maphash.Hash).Sum64())
	data = binary.BigEndian.AppendUint64(data[:], new(maphash.Hash).Sum64())
	data = binary.BigEndian.AppendUint64(data[:], new(maphash.Hash).Sum64())
	return trace.TraceID(data[:16]), trace.SpanID(data[16:])
}

func (g *lockFreeIdGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	data := make([]byte, 0, 8)
	data = binary.BigEndian.AppendUint64(data[:], new(maphash.Hash).Sum64())
	return trace.SpanID(data[:])
}
