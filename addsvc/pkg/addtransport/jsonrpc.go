package addtransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
	tjsonrpc "github.com/RangelReale/go-kit-typed/transport/jsonrpc"
	"github.com/go-kit/examples/addsvc/pkg/addendpoint"
	"github.com/go-kit/examples/addsvc/pkg/addservice"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport/http/jsonrpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
)

// NewJSONRPCHandler returns a JSON RPC Server/Handler that can be passed to http.Handle()
func NewJSONRPCHandler(endpoints addendpoint.Set, logger log.Logger) *jsonrpc.Server {
	handler := jsonrpc.NewServer(
		makeEndpointCodecMap(endpoints),
		jsonrpc.ServerErrorLogger(logger),
	)
	return handler
}

// NewJSONRPCClient returns an addservice backed by a JSON RPC over HTTP server
// living at the remote instance. We expect instance to come from a service
// discovery system, so likely of the form "host:port". We bake-in certain
// middlewares, implementing the client library pattern.
func NewJSONRPCClient(instance string, tracer stdopentracing.Tracer, logger log.Logger) (addservice.Service, error) {
	// Quickly sanitize the instance string.
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	var sumEndpoint tendpoint.Endpoint[addendpoint.SumRequest, addendpoint.SumResponse]
	{
		sumEndpoint = tjsonrpc.NewClient[addendpoint.SumRequest, addendpoint.SumResponse](
			u,
			"sum",
			tjsonrpc.ClientRequestEncoder(encodeSumRequest),
			tjsonrpc.ClientResponseDecoder(decodeSumResponse),
		).Endpoint()
		sumEndpoint = tendpoint.MiddlewareWrapper(opentracing.TraceClient(tracer, "Sum"), sumEndpoint)
		sumEndpoint = tendpoint.MiddlewareWrapper(limiter, sumEndpoint)
		sumEndpoint = tendpoint.MiddlewareWrapper(circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Sum",
			Timeout: 30 * time.Second,
		})), sumEndpoint)
	}

	var concatEndpoint tendpoint.Endpoint[addendpoint.ConcatRequest, addendpoint.ConcatResponse]
	{
		concatEndpoint = tjsonrpc.NewClient[addendpoint.ConcatRequest, addendpoint.ConcatResponse](
			u,
			"concat",
			tjsonrpc.ClientRequestEncoder(encodeConcatRequest),
			tjsonrpc.ClientResponseDecoder(decodeConcatResponse),
		).Endpoint()
		concatEndpoint = tendpoint.MiddlewareWrapper(opentracing.TraceClient(tracer, "Concat"), concatEndpoint)
		concatEndpoint = tendpoint.MiddlewareWrapper(limiter, concatEndpoint)
		concatEndpoint = tendpoint.MiddlewareWrapper(circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Concat",
			Timeout: 30 * time.Second,
		})), concatEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return addendpoint.Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}, nil

}

// makeEndpointCodecMap returns a codec map configured for the addsvc.
func makeEndpointCodecMap(endpoints addendpoint.Set) jsonrpc.EndpointCodecMap {
	return jsonrpc.EndpointCodecMap{
		"sum": tjsonrpc.EndpointCodecReverseAdapter(tjsonrpc.EndpointCodec[addendpoint.SumRequest, addendpoint.SumResponse]{
			Endpoint: endpoints.SumEndpoint,
			Decode:   decodeSumRequest,
			Encode:   encodeSumResponse,
		}),
		"concat": tjsonrpc.EndpointCodecReverseAdapter(tjsonrpc.EndpointCodec[addendpoint.ConcatRequest, addendpoint.ConcatResponse]{
			Endpoint: endpoints.ConcatEndpoint,
			Decode:   decodeConcatRequest,
			Encode:   encodeConcatResponse,
		}),
	}
}

func decodeSumRequest(_ context.Context, msg json.RawMessage) (addendpoint.SumRequest, error) {
	var req addendpoint.SumRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return addendpoint.SumRequest{}, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to sum request: %s", err),
		}
	}
	return req, nil
}

func encodeSumResponse(_ context.Context, res addendpoint.SumResponse) (json.RawMessage, error) {
	// res, ok := obj.(addendpoint.SumResponse)
	// if !ok {
	// 	return nil, &jsonrpc.Error{
	// 		Code:    -32000,
	// 		Message: fmt.Sprintf("Asserting result to *SumResponse failed. Got %T, %+v", obj, obj),
	// 	}
	// }
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal response: %s", err)
	}
	return b, nil
}

func decodeSumResponse(_ context.Context, res jsonrpc.Response) (addendpoint.SumResponse, error) {
	if res.Error != nil {
		return addendpoint.SumResponse{}, *res.Error
	}
	var sumres addendpoint.SumResponse
	err := json.Unmarshal(res.Result, &sumres)
	if err != nil {
		return addendpoint.SumResponse{}, fmt.Errorf("couldn't unmarshal body to SumResponse: %s", err)
	}
	return sumres, nil
}

func encodeSumRequest(_ context.Context, req addendpoint.SumRequest) (json.RawMessage, error) {
	// req, ok := obj.(addendpoint.SumRequest)
	// if !ok {
	// 	return nil, fmt.Errorf("couldn't assert request as SumRequest, got %T", obj)
	// }
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal request: %s", err)
	}
	return b, nil
}

func decodeConcatRequest(_ context.Context, msg json.RawMessage) (addendpoint.ConcatRequest, error) {
	var req addendpoint.ConcatRequest
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return addendpoint.ConcatRequest{}, &jsonrpc.Error{
			Code:    -32000,
			Message: fmt.Sprintf("couldn't unmarshal body to concat request: %s", err),
		}
	}
	return req, nil
}

func encodeConcatResponse(_ context.Context, res addendpoint.ConcatResponse) (json.RawMessage, error) {
	// res, ok := obj.(addendpoint.ConcatResponse)
	// if !ok {
	// 	return nil, &jsonrpc.Error{
	// 		Code:    -32000,
	// 		Message: fmt.Sprintf("Asserting result to *ConcatResponse failed. Got %T, %+v", obj, obj),
	// 	}
	// }
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal response: %s", err)
	}
	return b, nil
}

func decodeConcatResponse(_ context.Context, res jsonrpc.Response) (addendpoint.ConcatResponse, error) {
	if res.Error != nil {
		return addendpoint.ConcatResponse{}, *res.Error
	}
	var concatres addendpoint.ConcatResponse
	err := json.Unmarshal(res.Result, &concatres)
	if err != nil {
		return addendpoint.ConcatResponse{}, fmt.Errorf("couldn't unmarshal body to ConcatResponse: %s", err)
	}
	return concatres, nil
}

func encodeConcatRequest(_ context.Context, req addendpoint.ConcatRequest) (json.RawMessage, error) {
	// req, ok := obj.(addendpoint.ConcatRequest)
	// if !ok {
	// 	return nil, fmt.Errorf("couldn't assert request as ConcatRequest, got %T", obj)
	// }
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal request: %s", err)
	}
	return b, nil
}
