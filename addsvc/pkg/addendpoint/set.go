package addendpoint

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
	tmiddleware "github.com/RangelReale/go-kit-typed/endpoint/middleware"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"

	"github.com/go-kit/examples/addsvc/pkg/addservice"
)

// Set collects all of the endpoints that compose an add service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
type Set struct {
	SumEndpoint    tendpoint.Endpoint[SumRequest, SumResponse]
	ConcatEndpoint tendpoint.Endpoint[ConcatRequest, ConcatResponse]
}

// New returns a Set that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(svc addservice.Service, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) Set {
	var sumEndpoint tendpoint.Endpoint[SumRequest, SumResponse]
	{
		sumEndpoint = MakeSumEndpoint(svc)
		// Sum is limited to 1 request per second with burst of 1 request.
		// Note, rate is defined as a time interval between requests.
		sumEndpoint = tmiddleware.Wrapper(ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1)), sumEndpoint)
		sumEndpoint = tmiddleware.Wrapper(circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{})), sumEndpoint)
		sumEndpoint = tmiddleware.Wrapper(opentracing.TraceServer(otTracer, "Sum"), sumEndpoint)
		if zipkinTracer != nil {
			sumEndpoint = tmiddleware.Wrapper(zipkin.TraceEndpoint(zipkinTracer, "Sum"), sumEndpoint)
		}
		sumEndpoint = tmiddleware.Wrapper(LoggingMiddleware(log.With(logger, "method", "Sum")), sumEndpoint)
		sumEndpoint = tmiddleware.Wrapper(InstrumentingMiddleware(duration.With("method", "Sum")), sumEndpoint)
	}
	var concatEndpoint tendpoint.Endpoint[ConcatRequest, ConcatResponse]
	{
		concatEndpoint = MakeConcatEndpoint(svc)
		// Concat is limited to 1 request per second with burst of 100 requests.
		// Note, rate is defined as a number of requests per second.
		concatEndpoint = tmiddleware.Wrapper(ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Limit(1), 100)), concatEndpoint)
		concatEndpoint = tmiddleware.Wrapper(circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{})), concatEndpoint)
		concatEndpoint = tmiddleware.Wrapper(opentracing.TraceServer(otTracer, "Concat"), concatEndpoint)
		if zipkinTracer != nil {
			concatEndpoint = tmiddleware.Wrapper(zipkin.TraceEndpoint(zipkinTracer, "Concat"), concatEndpoint)
		}
		concatEndpoint = tmiddleware.Wrapper(LoggingMiddleware(log.With(logger, "method", "Concat")), concatEndpoint)
		concatEndpoint = tmiddleware.Wrapper(InstrumentingMiddleware(duration.With("method", "Concat")), concatEndpoint)
	}
	return Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}

// Sum implements the service interface, so Set may be used as a service.
// This is primarily useful in the context of a client library.
func (s Set) Sum(ctx context.Context, a, b int) (int, error) {
	response, err := s.SumEndpoint(ctx, SumRequest{A: a, B: b})
	if err != nil {
		return 0, err
	}
	return response.V, response.Err
}

// Concat implements the service interface, so Set may be used as a
// service. This is primarily useful in the context of a client library.
func (s Set) Concat(ctx context.Context, a, b string) (string, error) {
	response, err := s.ConcatEndpoint(ctx, ConcatRequest{A: a, B: b})
	if err != nil {
		return "", err
	}
	return response.V, response.Err
}

// MakeSumEndpoint constructs a Sum endpoint wrapping the service.
func MakeSumEndpoint(s addservice.Service) tendpoint.Endpoint[SumRequest, SumResponse] {
	return func(ctx context.Context, req SumRequest) (response SumResponse, err error) {
		v, err := s.Sum(ctx, req.A, req.B)
		return SumResponse{V: v, Err: err}, nil
	}
}

// MakeConcatEndpoint constructs a Concat endpoint wrapping the service.
func MakeConcatEndpoint(s addservice.Service) tendpoint.Endpoint[ConcatRequest, ConcatResponse] {
	return func(ctx context.Context, req ConcatRequest) (response ConcatResponse, err error) {
		v, err := s.Concat(ctx, req.A, req.B)
		return ConcatResponse{V: v, Err: err}, nil
	}
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = SumResponse{}
	_ endpoint.Failer = ConcatResponse{}
)

// SumRequest collects the request parameters for the Sum method.
type SumRequest struct {
	A, B int
}

// SumResponse collects the response values for the Sum method.
type SumResponse struct {
	V   int   `json:"v"`
	Err error `json:"-"` // should be intercepted by Failed/errorEncoder
}

// Failed implements endpoint.Failer.
func (r SumResponse) Failed() error { return r.Err }

// ConcatRequest collects the request parameters for the Concat method.
type ConcatRequest struct {
	A, B string
}

// ConcatResponse collects the response values for the Concat method.
type ConcatResponse struct {
	V   string `json:"v"`
	Err error  `json:"-"`
}

// Failed implements endpoint.Failer.
func (r ConcatResponse) Failed() error { return r.Err }
