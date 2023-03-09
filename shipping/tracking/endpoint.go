package tracking

import (
	"context"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
)

type trackCargoRequest struct {
	ID string
}

type trackCargoResponse struct {
	Cargo *Cargo `json:"cargo,omitempty"`
	Err   error  `json:"error,omitempty"`
}

func (r trackCargoResponse) error() error { return r.Err }

func makeTrackCargoEndpoint(ts Service) tendpoint.Endpoint[trackCargoRequest, trackCargoResponse] {
	return func(ctx context.Context, req trackCargoRequest) (trackCargoResponse, error) {
		c, err := ts.Track(req.ID)
		return trackCargoResponse{Cargo: &c, Err: err}, nil
	}
}
