package booking

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	thttptransport "github.com/RangelReale/go-kit-typed/transport/http"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/examples/shipping/cargo"
	"github.com/go-kit/examples/shipping/location"
)

// MakeHandler returns a handler for the booking service.
func MakeHandler(bs Service, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	bookCargoHandler := thttptransport.NewServer(
		makeBookCargoEndpoint(bs),
		decodeBookCargoRequest,
		thttptransport.EncodeResponseFuncAdapter[bookCargoResponse](encodeResponse),
		opts...,
	)
	loadCargoHandler := thttptransport.NewServer(
		makeLoadCargoEndpoint(bs),
		decodeLoadCargoRequest,
		thttptransport.EncodeResponseFuncAdapter[loadCargoResponse](encodeResponse),
		opts...,
	)
	requestRoutesHandler := thttptransport.NewServer(
		makeRequestRoutesEndpoint(bs),
		decodeRequestRoutesRequest,
		thttptransport.EncodeResponseFuncAdapter[requestRoutesResponse](encodeResponse),
		opts...,
	)
	assignToRouteHandler := thttptransport.NewServer(
		makeAssignToRouteEndpoint(bs),
		decodeAssignToRouteRequest,
		thttptransport.EncodeResponseFuncAdapter[assignToRouteResponse](encodeResponse),
		opts...,
	)
	changeDestinationHandler := thttptransport.NewServer(
		makeChangeDestinationEndpoint(bs),
		decodeChangeDestinationRequest,
		thttptransport.EncodeResponseFuncAdapter[changeDestinationResponse](encodeResponse),
		opts...,
	)
	listCargosHandler := thttptransport.NewServer(
		makeListCargosEndpoint(bs),
		decodeListCargosRequest,
		thttptransport.EncodeResponseFuncAdapter[listCargosResponse](encodeResponse),
		opts...,
	)
	listLocationsHandler := thttptransport.NewServer(
		makeListLocationsEndpoint(bs),
		decodeListLocationsRequest,
		thttptransport.EncodeResponseFuncAdapter[listLocationsResponse](encodeResponse),
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/booking/v1/cargos", bookCargoHandler).Methods("POST")
	r.Handle("/booking/v1/cargos", listCargosHandler).Methods("GET")
	r.Handle("/booking/v1/cargos/{id}", loadCargoHandler).Methods("GET")
	r.Handle("/booking/v1/cargos/{id}/request_routes", requestRoutesHandler).Methods("GET")
	r.Handle("/booking/v1/cargos/{id}/assign_to_route", assignToRouteHandler).Methods("POST")
	r.Handle("/booking/v1/cargos/{id}/change_destination", changeDestinationHandler).Methods("POST")
	r.Handle("/booking/v1/locations", listLocationsHandler).Methods("GET")

	return r
}

var errBadRoute = errors.New("bad route")

func decodeBookCargoRequest(_ context.Context, r *http.Request) (bookCargoRequest, error) {
	var body struct {
		Origin          string    `json:"origin"`
		Destination     string    `json:"destination"`
		ArrivalDeadline time.Time `json:"arrival_deadline"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return bookCargoRequest{}, err
	}

	return bookCargoRequest{
		Origin:          location.UNLocode(body.Origin),
		Destination:     location.UNLocode(body.Destination),
		ArrivalDeadline: body.ArrivalDeadline,
	}, nil
}

func decodeLoadCargoRequest(_ context.Context, r *http.Request) (loadCargoRequest, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return loadCargoRequest{}, errBadRoute
	}
	return loadCargoRequest{ID: cargo.TrackingID(id)}, nil
}

func decodeRequestRoutesRequest(_ context.Context, r *http.Request) (requestRoutesRequest, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return requestRoutesRequest{}, errBadRoute
	}
	return requestRoutesRequest{ID: cargo.TrackingID(id)}, nil
}

func decodeAssignToRouteRequest(_ context.Context, r *http.Request) (assignToRouteRequest, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return assignToRouteRequest{}, errBadRoute
	}

	var itinerary cargo.Itinerary
	if err := json.NewDecoder(r.Body).Decode(&itinerary); err != nil {
		return assignToRouteRequest{}, err
	}

	return assignToRouteRequest{
		ID:        cargo.TrackingID(id),
		Itinerary: itinerary,
	}, nil
}

func decodeChangeDestinationRequest(_ context.Context, r *http.Request) (changeDestinationRequest, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return changeDestinationRequest{}, errBadRoute
	}

	var body struct {
		Destination string `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return changeDestinationRequest{}, err
	}

	return changeDestinationRequest{
		ID:          cargo.TrackingID(id),
		Destination: location.UNLocode(body.Destination),
	}, nil
}

func decodeListCargosRequest(_ context.Context, r *http.Request) (listCargosRequest, error) {
	return listCargosRequest{}, nil
}

func decodeListLocationsRequest(_ context.Context, r *http.Request) (listLocationsRequest, error) {
	return listLocationsRequest{}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case cargo.ErrUnknown:
		w.WriteHeader(http.StatusNotFound)
	case ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
