package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
	thttptransport "github.com/RangelReale/go-kit-typed/transport/http"
	tnatstransport "github.com/RangelReale/go-kit-typed/transport/nats"
	httptransport "github.com/go-kit/kit/transport/http"
	natstransport "github.com/go-kit/kit/transport/nats"

	"github.com/nats-io/nats.go"
)

// StringService provides operations on strings.
type StringService interface {
	Uppercase(context.Context, string) (string, error)
	Count(context.Context, string) int
}

// stringService is a concrete implementation of StringService
type stringService struct{}

func (stringService) Uppercase(_ context.Context, s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	return strings.ToUpper(s), nil
}

func (stringService) Count(_ context.Context, s string) int {
	return len(s)
}

// ErrEmpty is returned when an input string is empty.
var ErrEmpty = errors.New("empty string")

// For each method, we define request and response structs
type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}

// Endpoints are a primary abstraction in go-kit. An endpoint represents a single RPC (method in our service interface)
func makeUppercaseHTTPEndpoint(nc *nats.Conn) tendpoint.Endpoint[uppercaseRequest, uppercaseResponse] {
	return tnatstransport.NewPublisherStdEnc[uppercaseRequest, uppercaseResponse](
		nc,
		"stringsvc.uppercase",
		natstransport.EncodeJSONRequest,
		decodeUppercaseResponse,
	).Endpoint()
}

func makeCountHTTPEndpoint(nc *nats.Conn) tendpoint.Endpoint[countRequest, countResponse] {
	return tnatstransport.NewPublisherStdEnc[countRequest, countResponse](
		nc,
		"stringsvc.count",
		natstransport.EncodeJSONRequest,
		decodeCountResponse,
	).Endpoint()
}

func makeUppercaseEndpoint(svc StringService) tendpoint.Endpoint[uppercaseRequest, uppercaseResponse] {
	return func(ctx context.Context, req uppercaseRequest) (uppercaseResponse, error) {
		v, err := svc.Uppercase(ctx, req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}
		return uppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc StringService) tendpoint.Endpoint[countRequest, countResponse] {
	return func(ctx context.Context, req countRequest) (countResponse, error) {
		v := svc.Count(ctx, req.S)
		return countResponse{v}, nil
	}
}

// Transports expose the service to the network. In this fourth example we utilize JSON over NATS and HTTP.
func main() {
	svc := stringService{}

	natsURL := flag.String("nats-url", nats.DefaultURL, "URL for connection to NATS")
	flag.Parse()

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	uppercaseHTTPHandler := thttptransport.NewServerStdEnc(
		makeUppercaseHTTPEndpoint(nc),
		decodeUppercaseHTTPRequest,
		httptransport.EncodeJSONResponse,
	)

	countHTTPHandler := thttptransport.NewServerStdEnc(
		makeCountHTTPEndpoint(nc),
		decodeCountHTTPRequest,
		httptransport.EncodeJSONResponse,
	)

	uppercaseHandler := tnatstransport.NewSubscriberStdEnc(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		natstransport.EncodeJSONResponse,
	)

	countHandler := tnatstransport.NewSubscriberStdEnc(
		makeCountEndpoint(svc),
		decodeCountRequest,
		natstransport.EncodeJSONResponse,
	)

	uSub, err := nc.QueueSubscribe("stringsvc.uppercase", "stringsvc", uppercaseHandler.ServeMsg(nc))
	if err != nil {
		log.Fatal(err)
	}
	defer uSub.Unsubscribe()

	cSub, err := nc.QueueSubscribe("stringsvc.count", "stringsvc", countHandler.ServeMsg(nc))
	if err != nil {
		log.Fatal(err)
	}
	defer cSub.Unsubscribe()

	http.Handle("/uppercase", uppercaseHTTPHandler)
	http.Handle("/count", countHTTPHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func decodeUppercaseHTTPRequest(_ context.Context, r *http.Request) (uppercaseRequest, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return uppercaseRequest{}, err
	}
	return request, nil
}

func decodeCountHTTPRequest(_ context.Context, r *http.Request) (countRequest, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return countRequest{}, err
	}
	return request, nil
}

func decodeUppercaseResponse(_ context.Context, msg *nats.Msg) (uppercaseResponse, error) {
	var response uppercaseResponse

	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return uppercaseResponse{}, err
	}

	return response, nil
}

func decodeCountResponse(_ context.Context, msg *nats.Msg) (countResponse, error) {
	var response countResponse

	if err := json.Unmarshal(msg.Data, &response); err != nil {
		return countResponse{}, err
	}

	return response, nil
}

func decodeUppercaseRequest(_ context.Context, msg *nats.Msg) (uppercaseRequest, error) {
	var request uppercaseRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return uppercaseRequest{}, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, msg *nats.Msg) (countRequest, error) {
	var request countRequest

	if err := json.Unmarshal(msg.Data, &request); err != nil {
		return countRequest{}, err
	}
	return request, nil
}
