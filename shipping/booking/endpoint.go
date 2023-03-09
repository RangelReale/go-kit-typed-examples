package booking

import (
	"context"
	"time"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
	"github.com/go-kit/examples/shipping/cargo"
	"github.com/go-kit/examples/shipping/location"
)

type bookCargoRequest struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

type bookCargoResponse struct {
	ID  cargo.TrackingID `json:"tracking_id,omitempty"`
	Err error            `json:"error,omitempty"`
}

func (r bookCargoResponse) error() error { return r.Err }

func makeBookCargoEndpoint(s Service) tendpoint.Endpoint[bookCargoRequest, bookCargoResponse] {
	return func(ctx context.Context, req bookCargoRequest) (bookCargoResponse, error) {
		id, err := s.BookNewCargo(req.Origin, req.Destination, req.ArrivalDeadline)
		return bookCargoResponse{ID: id, Err: err}, nil
	}
}

type loadCargoRequest struct {
	ID cargo.TrackingID
}

type loadCargoResponse struct {
	Cargo *Cargo `json:"cargo,omitempty"`
	Err   error  `json:"error,omitempty"`
}

func (r loadCargoResponse) error() error { return r.Err }

func makeLoadCargoEndpoint(s Service) tendpoint.Endpoint[loadCargoRequest, loadCargoResponse] {
	return func(ctx context.Context, req loadCargoRequest) (loadCargoResponse, error) {
		c, err := s.LoadCargo(req.ID)
		return loadCargoResponse{Cargo: &c, Err: err}, nil
	}
}

type requestRoutesRequest struct {
	ID cargo.TrackingID
}

type requestRoutesResponse struct {
	Routes []cargo.Itinerary `json:"routes,omitempty"`
	Err    error             `json:"error,omitempty"`
}

func (r requestRoutesResponse) error() error { return r.Err }

func makeRequestRoutesEndpoint(s Service) tendpoint.Endpoint[requestRoutesRequest, requestRoutesResponse] {
	return func(ctx context.Context, req requestRoutesRequest) (requestRoutesResponse, error) {
		itin := s.RequestPossibleRoutesForCargo(req.ID)
		return requestRoutesResponse{Routes: itin, Err: nil}, nil
	}
}

type assignToRouteRequest struct {
	ID        cargo.TrackingID
	Itinerary cargo.Itinerary
}

type assignToRouteResponse struct {
	Err error `json:"error,omitempty"`
}

func (r assignToRouteResponse) error() error { return r.Err }

func makeAssignToRouteEndpoint(s Service) tendpoint.Endpoint[assignToRouteRequest, assignToRouteResponse] {
	return func(ctx context.Context, req assignToRouteRequest) (assignToRouteResponse, error) {
		err := s.AssignCargoToRoute(req.ID, req.Itinerary)
		return assignToRouteResponse{Err: err}, nil
	}
}

type changeDestinationRequest struct {
	ID          cargo.TrackingID
	Destination location.UNLocode
}

type changeDestinationResponse struct {
	Err error `json:"error,omitempty"`
}

func (r changeDestinationResponse) error() error { return r.Err }

func makeChangeDestinationEndpoint(s Service) tendpoint.Endpoint[changeDestinationRequest, changeDestinationResponse] {
	return func(ctx context.Context, req changeDestinationRequest) (changeDestinationResponse, error) {
		err := s.ChangeDestination(req.ID, req.Destination)
		return changeDestinationResponse{Err: err}, nil
	}
}

type listCargosRequest struct{}

type listCargosResponse struct {
	Cargos []Cargo `json:"cargos,omitempty"`
	Err    error   `json:"error,omitempty"`
}

func (r listCargosResponse) error() error { return r.Err }

func makeListCargosEndpoint(s Service) tendpoint.Endpoint[listCargosRequest, listCargosResponse] {
	return func(ctx context.Context, req listCargosRequest) (listCargosResponse, error) {
		return listCargosResponse{Cargos: s.Cargos(), Err: nil}, nil
	}
}

type listLocationsRequest struct {
}

type listLocationsResponse struct {
	Locations []Location `json:"locations,omitempty"`
	Err       error      `json:"error,omitempty"`
}

func makeListLocationsEndpoint(s Service) tendpoint.Endpoint[listLocationsRequest, listLocationsResponse] {
	return func(ctx context.Context, req listLocationsRequest) (listLocationsResponse, error) {
		return listLocationsResponse{Locations: s.Locations(), Err: nil}, nil
	}
}
