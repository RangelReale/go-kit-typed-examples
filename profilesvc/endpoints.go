package profilesvc

import (
	"context"
	"net/url"
	"strings"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
	thttptransport "github.com/RangelReale/go-kit-typed/transport/http"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them
// into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	PostProfileEndpoint   tendpoint.Endpoint[postProfileRequest, postProfileResponse]
	GetProfileEndpoint    tendpoint.Endpoint[getProfileRequest, getProfileResponse]
	PutProfileEndpoint    tendpoint.Endpoint[putProfileRequest, putProfileResponse]
	PatchProfileEndpoint  tendpoint.Endpoint[patchProfileRequest, patchProfileResponse]
	DeleteProfileEndpoint tendpoint.Endpoint[deleteProfileRequest, deleteProfileResponse]
	GetAddressesEndpoint  tendpoint.Endpoint[getAddressesRequest, getAddressesResponse]
	GetAddressEndpoint    tendpoint.Endpoint[getAddressRequest, getAddressResponse]
	PostAddressEndpoint   tendpoint.Endpoint[postAddressRequest, postAddressResponse]
	DeleteAddressEndpoint tendpoint.Endpoint[deleteAddressRequest, deleteAddressResponse]
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service. Useful in a profilesvc
// server.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		PostProfileEndpoint:   MakePostProfileEndpoint(s),
		GetProfileEndpoint:    MakeGetProfileEndpoint(s),
		PutProfileEndpoint:    MakePutProfileEndpoint(s),
		PatchProfileEndpoint:  MakePatchProfileEndpoint(s),
		DeleteProfileEndpoint: MakeDeleteProfileEndpoint(s),
		GetAddressesEndpoint:  MakeGetAddressesEndpoint(s),
		GetAddressEndpoint:    MakeGetAddressEndpoint(s),
		PostAddressEndpoint:   MakePostAddressEndpoint(s),
		DeleteAddressEndpoint: MakeDeleteAddressEndpoint(s),
	}
}

// MakeClientEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the remote instance, via a transport/http.Client.
// Useful in a profilesvc client.
func MakeClientEndpoints(instance string) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	options := []httptransport.ClientOption{}

	// Note that the request encoders need to modify the request URL, changing
	// the path. That's fine: we simply need to provide specific encoders for
	// each endpoint.

	return Endpoints{
		PostProfileEndpoint:   thttptransport.NewClient("POST", tgt, encodePostProfileRequest, decodePostProfileResponse, options...).Endpoint(),
		GetProfileEndpoint:    thttptransport.NewClient("GET", tgt, encodeGetProfileRequest, decodeGetProfileResponse, options...).Endpoint(),
		PutProfileEndpoint:    thttptransport.NewClient("PUT", tgt, encodePutProfileRequest, decodePutProfileResponse, options...).Endpoint(),
		PatchProfileEndpoint:  thttptransport.NewClient("PATCH", tgt, encodePatchProfileRequest, decodePatchProfileResponse, options...).Endpoint(),
		DeleteProfileEndpoint: thttptransport.NewClient("DELETE", tgt, encodeDeleteProfileRequest, decodeDeleteProfileResponse, options...).Endpoint(),
		GetAddressesEndpoint:  thttptransport.NewClient("GET", tgt, encodeGetAddressesRequest, decodeGetAddressesResponse, options...).Endpoint(),
		GetAddressEndpoint:    thttptransport.NewClient("GET", tgt, encodeGetAddressRequest, decodeGetAddressResponse, options...).Endpoint(),
		PostAddressEndpoint:   thttptransport.NewClient("POST", tgt, encodePostAddressRequest, decodePostAddressResponse, options...).Endpoint(),
		DeleteAddressEndpoint: thttptransport.NewClient("DELETE", tgt, encodeDeleteAddressRequest, decodeDeleteAddressResponse, options...).Endpoint(),
	}, nil
}

// PostProfile implements Service. Primarily useful in a client.
func (e Endpoints) PostProfile(ctx context.Context, p Profile) error {
	request := postProfileRequest{Profile: p}
	resp, err := e.PostProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	return resp.Err
}

// GetProfile implements Service. Primarily useful in a client.
func (e Endpoints) GetProfile(ctx context.Context, id string) (Profile, error) {
	request := getProfileRequest{ID: id}
	resp, err := e.GetProfileEndpoint(ctx, request)
	if err != nil {
		return Profile{}, err
	}
	return resp.Profile, resp.Err
}

// PutProfile implements Service. Primarily useful in a client.
func (e Endpoints) PutProfile(ctx context.Context, id string, p Profile) error {
	request := putProfileRequest{ID: id, Profile: p}
	resp, err := e.PutProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	return resp.Err
}

// PatchProfile implements Service. Primarily useful in a client.
func (e Endpoints) PatchProfile(ctx context.Context, id string, p Profile) error {
	request := patchProfileRequest{ID: id, Profile: p}
	resp, err := e.PatchProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	return resp.Err
}

// DeleteProfile implements Service. Primarily useful in a client.
func (e Endpoints) DeleteProfile(ctx context.Context, id string) error {
	request := deleteProfileRequest{ID: id}
	resp, err := e.DeleteProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	return resp.Err
}

// GetAddresses implements Service. Primarily useful in a client.
func (e Endpoints) GetAddresses(ctx context.Context, profileID string) ([]Address, error) {
	request := getAddressesRequest{ProfileID: profileID}
	resp, err := e.GetAddressesEndpoint(ctx, request)
	if err != nil {
		return nil, err
	}
	return resp.Addresses, resp.Err
}

// GetAddress implements Service. Primarily useful in a client.
func (e Endpoints) GetAddress(ctx context.Context, profileID string, addressID string) (Address, error) {
	request := getAddressRequest{ProfileID: profileID, AddressID: addressID}
	resp, err := e.GetAddressEndpoint(ctx, request)
	if err != nil {
		return Address{}, err
	}
	return resp.Address, resp.Err
}

// PostAddress implements Service. Primarily useful in a client.
func (e Endpoints) PostAddress(ctx context.Context, profileID string, a Address) error {
	request := postAddressRequest{ProfileID: profileID, Address: a}
	resp, err := e.PostAddressEndpoint(ctx, request)
	if err != nil {
		return err
	}
	return resp.Err
}

// DeleteAddress implements Service. Primarily useful in a client.
func (e Endpoints) DeleteAddress(ctx context.Context, profileID string, addressID string) error {
	request := deleteAddressRequest{ProfileID: profileID, AddressID: addressID}
	resp, err := e.DeleteAddressEndpoint(ctx, request)
	if err != nil {
		return err
	}
	return resp.Err
}

// MakePostProfileEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakePostProfileEndpoint(s Service) tendpoint.Endpoint[postProfileRequest, postProfileResponse] {
	return func(ctx context.Context, req postProfileRequest) (response postProfileResponse, err error) {
		e := s.PostProfile(ctx, req.Profile)
		return postProfileResponse{Err: e}, nil
	}
}

// MakeGetProfileEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetProfileEndpoint(s Service) tendpoint.Endpoint[getProfileRequest, getProfileResponse] {
	return func(ctx context.Context, req getProfileRequest) (response getProfileResponse, err error) {
		p, e := s.GetProfile(ctx, req.ID)
		return getProfileResponse{Profile: p, Err: e}, nil
	}
}

// MakePutProfileEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakePutProfileEndpoint(s Service) tendpoint.Endpoint[putProfileRequest, putProfileResponse] {
	return func(ctx context.Context, req putProfileRequest) (response putProfileResponse, err error) {
		e := s.PutProfile(ctx, req.ID, req.Profile)
		return putProfileResponse{Err: e}, nil
	}
}

// MakePatchProfileEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakePatchProfileEndpoint(s Service) tendpoint.Endpoint[patchProfileRequest, patchProfileResponse] {
	return func(ctx context.Context, req patchProfileRequest) (response patchProfileResponse, err error) {
		e := s.PatchProfile(ctx, req.ID, req.Profile)
		return patchProfileResponse{Err: e}, nil
	}
}

// MakeDeleteProfileEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeDeleteProfileEndpoint(s Service) tendpoint.Endpoint[deleteProfileRequest, deleteProfileResponse] {
	return func(ctx context.Context, req deleteProfileRequest) (response deleteProfileResponse, err error) {
		e := s.DeleteProfile(ctx, req.ID)
		return deleteProfileResponse{Err: e}, nil
	}
}

// MakeGetAddressesEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetAddressesEndpoint(s Service) tendpoint.Endpoint[getAddressesRequest, getAddressesResponse] {
	return func(ctx context.Context, req getAddressesRequest) (response getAddressesResponse, err error) {
		a, e := s.GetAddresses(ctx, req.ProfileID)
		return getAddressesResponse{Addresses: a, Err: e}, nil
	}
}

// MakeGetAddressEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetAddressEndpoint(s Service) tendpoint.Endpoint[getAddressRequest, getAddressResponse] {
	return func(ctx context.Context, req getAddressRequest) (response getAddressResponse, err error) {
		a, e := s.GetAddress(ctx, req.ProfileID, req.AddressID)
		return getAddressResponse{Address: a, Err: e}, nil
	}
}

// MakePostAddressEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakePostAddressEndpoint(s Service) tendpoint.Endpoint[postAddressRequest, postAddressResponse] {
	return func(ctx context.Context, req postAddressRequest) (response postAddressResponse, err error) {
		e := s.PostAddress(ctx, req.ProfileID, req.Address)
		return postAddressResponse{Err: e}, nil
	}
}

// MakeDeleteAddressEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeDeleteAddressEndpoint(s Service) tendpoint.Endpoint[deleteAddressRequest, deleteAddressResponse] {
	return func(ctx context.Context, req deleteAddressRequest) (response deleteAddressResponse, err error) {
		e := s.DeleteAddress(ctx, req.ProfileID, req.AddressID)
		return deleteAddressResponse{Err: e}, nil
	}
}

// We have two options to return errors from the business logic.
//
// We could return the error via the endpoint itself. That makes certain things
// a little bit easier, like providing non-200 HTTP responses to the client. But
// Go kit assumes that endpoint errors are (or may be treated as)
// transport-domain errors. For example, an endpoint error will count against a
// circuit breaker error count.
//
// Therefore, it's often better to return service (business logic) errors in the
// response object. This means we have to do a bit more work in the HTTP
// response encoder to detect e.g. a not-found error and provide a proper HTTP
// status code. That work is done with the errorer interface, in transport.go.
// Response types that may contain business-logic errors implement that
// interface.

type postProfileRequest struct {
	Profile Profile
}

type postProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r postProfileResponse) error() error { return r.Err }

type getProfileRequest struct {
	ID string
}

type getProfileResponse struct {
	Profile Profile `json:"profile,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r getProfileResponse) error() error { return r.Err }

type putProfileRequest struct {
	ID      string
	Profile Profile
}

type putProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r putProfileResponse) error() error { return nil }

type patchProfileRequest struct {
	ID      string
	Profile Profile
}

type patchProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r patchProfileResponse) error() error { return r.Err }

type deleteProfileRequest struct {
	ID string
}

type deleteProfileResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteProfileResponse) error() error { return r.Err }

type getAddressesRequest struct {
	ProfileID string
}

type getAddressesResponse struct {
	Addresses []Address `json:"addresses,omitempty"`
	Err       error     `json:"err,omitempty"`
}

func (r getAddressesResponse) error() error { return r.Err }

type getAddressRequest struct {
	ProfileID string
	AddressID string
}

type getAddressResponse struct {
	Address Address `json:"address,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r getAddressResponse) error() error { return r.Err }

type postAddressRequest struct {
	ProfileID string
	Address   Address
}

type postAddressResponse struct {
	Err error `json:"err,omitempty"`
}

func (r postAddressResponse) error() error { return r.Err }

type deleteAddressRequest struct {
	ProfileID string
	AddressID string
}

type deleteAddressResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteAddressResponse) error() error { return r.Err }
