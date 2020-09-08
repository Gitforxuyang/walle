package trace

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc/metadata"
	"strings"

	"github.com/opentracing/opentracing-go"
)

// metadataReaderWriter satisfies both the opentracing.TextMapReader and
// opentracing.TextMapWriter interfaces.
type metadataReaderWriter struct {
	metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Tracer) StartServerSpanFromContext(ctx *gin.Context, name string, opts ...opentracing.StartSpanOption) (opentracing.Span, error) {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	} else if spanCtx, err := m.tracer.Extract(opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(ctx.Request.Header)); err == nil {
		opts = append(opts, opentracing.ChildOf(spanCtx))
	}

	sp := m.tracer.StartSpan(name, opts...)
	withTraceId(ctx, sp)
	ext.SpanKindRPCServer.Set(sp)
	ext.Component.Set(sp, "http")
	ext.HTTPMethod.Set(sp, ctx.Request.Method)
	//ctx = opentracing.ContextWithSpan(ctx, sp)
	return sp, nil
}

//func (m *Tracer) StartGRpcClientSpanFromContext(ctx context.Context, name string, opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span, error) {
//	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
//		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
//	}
//	sp := m.tracer.StartSpan(name, opts...)
//	//ctx = withTraceId(ctx, sp)
//	ext.SpanKindRPCClient.Set(sp)
//	ext.Component.Set(sp, "grpc")
//	md, ok := metadata.FromOutgoingContext(ctx)
//	if !ok {
//		md = metadata.New(nil)
//	} else {
//		md = md.Copy()
//	}
//	mdWriter := metadataReaderWriter{md}
//	err := m.tracer.Inject(sp.Context(), opentracing.TextMap, mdWriter)
//	if err != nil {
//		return nil, nil, err
//	}
//	ctx = metadata.NewOutgoingContext(ctx, md)
//	return ctx, sp, nil
//}
//
//func (m *Tracer) StartHttpClientSpanFromContext(ctx context.Context, name string, opts ...opentracing.StartSpanOption) (context.Context, opentracing.Span, error) {
//	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
//		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
//	}
//	sp := m.tracer.StartSpan(name, opts...)
//	//ctx = withTraceId(ctx, sp)
//	ext.SpanKindRPCClient.Set(sp)
//	ext.Component.Set(sp, "http")
//	md, ok := metadata.FromOutgoingContext(ctx)
//	if !ok {
//		md = metadata.New(nil)
//	} else {
//		md = md.Copy()
//	}
//	mdWriter := metadataReaderWriter{md}
//	err := m.tracer.Inject(sp.Context(), opentracing.HTTPHeaders, mdWriter)
//	if err != nil {
//		return nil, nil, err
//	}
//	ctx = metadata.NewOutgoingContext(ctx, md)
//	return ctx, sp, nil
//}

func withTraceId(ctx *gin.Context, span opentracing.Span) {
	s, ok := span.Context().(jaeger.SpanContext)
	if ok {
		ctx.Set("traceId", s.TraceID().String())
	}
}
