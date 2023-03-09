package handling

import (
	"context"
	"time"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"

	"github.com/go-kit/examples/shipping/cargo"
	"github.com/go-kit/examples/shipping/location"
	"github.com/go-kit/examples/shipping/voyage"
)

type registerIncidentRequest struct {
	ID             cargo.TrackingID
	Location       location.UNLocode
	Voyage         voyage.Number
	EventType      cargo.HandlingEventType
	CompletionTime time.Time
}

type registerIncidentResponse struct {
	Err error `json:"error,omitempty"`
}

func (r registerIncidentResponse) error() error { return r.Err }

func makeRegisterIncidentEndpoint(hs Service) tendpoint.Endpoint[registerIncidentRequest, registerIncidentResponse] {
	return func(ctx context.Context, req registerIncidentRequest) (registerIncidentResponse, error) {
		err := hs.RegisterHandlingEvent(req.CompletionTime, req.ID, req.Voyage, req.Location, req.EventType)
		return registerIncidentResponse{Err: err}, nil
	}
}
