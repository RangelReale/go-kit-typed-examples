package profilesvc

// The profilesvc is just over HTTP, so we just have a single transport.go.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	thttptransport "github.com/RangelReale/go-kit-typed/transport/http"
	"github.com/gorilla/mux"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a profilesvc server.
func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// POST    /profiles/                          adds another profile
	// GET     /profiles/:id                       retrieves the given profile by id
	// PUT     /profiles/:id                       post updated profile information about the profile
	// PATCH   /profiles/:id                       partial updated profile information
	// DELETE  /profiles/:id                       remove the given profile
	// GET     /profiles/:id/addresses/            retrieve addresses associated with the profile
	// GET     /profiles/:id/addresses/:addressID  retrieve a particular profile address
	// POST    /profiles/:id/addresses/            add a new address
	// DELETE  /profiles/:id/addresses/:addressID  remove an address

	r.Methods("POST").Path("/profiles/").Handler(thttptransport.NewServer(
		e.PostProfileEndpoint,
		decodePostProfileRequest,
		thttptransport.EncodeResponseFuncAdapter[postProfileResponse](encodeResponse),
		options...,
	))
	r.Methods("GET").Path("/profiles/{id}").Handler(thttptransport.NewServer(
		e.GetProfileEndpoint,
		decodeGetProfileRequest,
		thttptransport.EncodeResponseFuncAdapter[getProfileResponse](encodeResponse),
		options...,
	))
	r.Methods("PUT").Path("/profiles/{id}").Handler(thttptransport.NewServer(
		e.PutProfileEndpoint,
		decodePutProfileRequest,
		thttptransport.EncodeResponseFuncAdapter[putProfileResponse](encodeResponse),
		options...,
	))
	r.Methods("PATCH").Path("/profiles/{id}").Handler(thttptransport.NewServer(
		e.PatchProfileEndpoint,
		decodePatchProfileRequest,
		thttptransport.EncodeResponseFuncAdapter[patchProfileResponse](encodeResponse),
		options...,
	))
	r.Methods("DELETE").Path("/profiles/{id}").Handler(thttptransport.NewServer(
		e.DeleteProfileEndpoint,
		decodeDeleteProfileRequest,
		thttptransport.EncodeResponseFuncAdapter[deleteProfileResponse](encodeResponse),
		options...,
	))
	r.Methods("GET").Path("/profiles/{id}/addresses/").Handler(thttptransport.NewServer(
		e.GetAddressesEndpoint,
		decodeGetAddressesRequest,
		thttptransport.EncodeResponseFuncAdapter[getAddressesResponse](encodeResponse),
		options...,
	))
	r.Methods("GET").Path("/profiles/{id}/addresses/{addressID}").Handler(thttptransport.NewServer(
		e.GetAddressEndpoint,
		decodeGetAddressRequest,
		thttptransport.EncodeResponseFuncAdapter[getAddressResponse](encodeResponse),
		options...,
	))
	r.Methods("POST").Path("/profiles/{id}/addresses/").Handler(thttptransport.NewServer(
		e.PostAddressEndpoint,
		decodePostAddressRequest,
		thttptransport.EncodeResponseFuncAdapter[postAddressResponse](encodeResponse),
		options...,
	))
	r.Methods("DELETE").Path("/profiles/{id}/addresses/{addressID}").Handler(thttptransport.NewServer(
		e.DeleteAddressEndpoint,
		decodeDeleteAddressRequest,
		thttptransport.EncodeResponseFuncAdapter[deleteAddressResponse](encodeResponse),
		options...,
	))
	return r
}

func decodePostProfileRequest(_ context.Context, r *http.Request) (request postProfileRequest, err error) {
	var req postProfileRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Profile); e != nil {
		return postProfileRequest{}, e
	}
	return req, nil
}

func decodeGetProfileRequest(_ context.Context, r *http.Request) (request getProfileRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return getProfileRequest{}, ErrBadRouting
	}
	return getProfileRequest{ID: id}, nil
}

func decodePutProfileRequest(_ context.Context, r *http.Request) (request putProfileRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return putProfileRequest{}, ErrBadRouting
	}
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		return putProfileRequest{}, err
	}
	return putProfileRequest{
		ID:      id,
		Profile: profile,
	}, nil
}

func decodePatchProfileRequest(_ context.Context, r *http.Request) (request patchProfileRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return patchProfileRequest{}, ErrBadRouting
	}
	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		return patchProfileRequest{}, err
	}
	return patchProfileRequest{
		ID:      id,
		Profile: profile,
	}, nil
}

func decodeDeleteProfileRequest(_ context.Context, r *http.Request) (request deleteProfileRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return deleteProfileRequest{}, ErrBadRouting
	}
	return deleteProfileRequest{ID: id}, nil
}

func decodeGetAddressesRequest(_ context.Context, r *http.Request) (request getAddressesRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return getAddressesRequest{}, ErrBadRouting
	}
	return getAddressesRequest{ProfileID: id}, nil
}

func decodeGetAddressRequest(_ context.Context, r *http.Request) (request getAddressRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return getAddressRequest{}, ErrBadRouting
	}
	addressID, ok := vars["addressID"]
	if !ok {
		return getAddressRequest{}, ErrBadRouting
	}
	return getAddressRequest{
		ProfileID: id,
		AddressID: addressID,
	}, nil
}

func decodePostAddressRequest(_ context.Context, r *http.Request) (request postAddressRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return postAddressRequest{}, ErrBadRouting
	}
	var address Address
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		return postAddressRequest{}, err
	}
	return postAddressRequest{
		ProfileID: id,
		Address:   address,
	}, nil
}

func decodeDeleteAddressRequest(_ context.Context, r *http.Request) (request deleteAddressRequest, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return deleteAddressRequest{}, ErrBadRouting
	}
	addressID, ok := vars["addressID"]
	if !ok {
		return deleteAddressRequest{}, ErrBadRouting
	}
	return deleteAddressRequest{
		ProfileID: id,
		AddressID: addressID,
	}, nil
}

func encodePostProfileRequest(ctx context.Context, req *http.Request, request postProfileRequest) error {
	// r.Methods("POST").Path("/profiles/")
	req.URL.Path = "/profiles/"
	return encodeRequest(ctx, req, request)
}

func encodeGetProfileRequest(ctx context.Context, req *http.Request, request getProfileRequest) error {
	// r.Methods("GET").Path("/profiles/{id}")
	profileID := url.QueryEscape(request.ID)
	req.URL.Path = "/profiles/" + profileID
	return encodeRequest(ctx, req, request)
}

func encodePutProfileRequest(ctx context.Context, req *http.Request, request putProfileRequest) error {
	// r.Methods("PUT").Path("/profiles/{id}")
	profileID := url.QueryEscape(request.ID)
	req.URL.Path = "/profiles/" + profileID
	return encodeRequest(ctx, req, request)
}

func encodePatchProfileRequest(ctx context.Context, req *http.Request, request patchProfileRequest) error {
	// r.Methods("PATCH").Path("/profiles/{id}")
	profileID := url.QueryEscape(request.ID)
	req.URL.Path = "/profiles/" + profileID
	return encodeRequest(ctx, req, request)
}

func encodeDeleteProfileRequest(ctx context.Context, req *http.Request, request deleteProfileRequest) error {
	// r.Methods("DELETE").Path("/profiles/{id}")
	profileID := url.QueryEscape(request.ID)
	req.URL.Path = "/profiles/" + profileID
	return encodeRequest(ctx, req, request)
}

func encodeGetAddressesRequest(ctx context.Context, req *http.Request, request getAddressesRequest) error {
	// r.Methods("GET").Path("/profiles/{id}/addresses/")
	profileID := url.QueryEscape(request.ProfileID)
	req.URL.Path = "/profiles/" + profileID + "/addresses/"
	return encodeRequest(ctx, req, request)
}

func encodeGetAddressRequest(ctx context.Context, req *http.Request, request getAddressRequest) error {
	// r.Methods("GET").Path("/profiles/{id}/addresses/{addressID}")
	profileID := url.QueryEscape(request.ProfileID)
	addressID := url.QueryEscape(request.AddressID)
	req.URL.Path = "/profiles/" + profileID + "/addresses/" + addressID
	return encodeRequest(ctx, req, request)
}

func encodePostAddressRequest(ctx context.Context, req *http.Request, request postAddressRequest) error {
	// r.Methods("POST").Path("/profiles/{id}/addresses/")
	profileID := url.QueryEscape(request.ProfileID)
	req.URL.Path = "/profiles/" + profileID + "/addresses/"
	return encodeRequest(ctx, req, request)
}

func encodeDeleteAddressRequest(ctx context.Context, req *http.Request, request deleteAddressRequest) error {
	// r.Methods("DELETE").Path("/profiles/{id}/addresses/{addressID}")
	profileID := url.QueryEscape(request.ProfileID)
	addressID := url.QueryEscape(request.AddressID)
	req.URL.Path = "/profiles/" + profileID + "/addresses/" + addressID
	return encodeRequest(ctx, req, request)
}

func decodePostProfileResponse(_ context.Context, resp *http.Response) (postProfileResponse, error) {
	var response postProfileResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeGetProfileResponse(_ context.Context, resp *http.Response) (getProfileResponse, error) {
	var response getProfileResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodePutProfileResponse(_ context.Context, resp *http.Response) (putProfileResponse, error) {
	var response putProfileResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodePatchProfileResponse(_ context.Context, resp *http.Response) (patchProfileResponse, error) {
	var response patchProfileResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeDeleteProfileResponse(_ context.Context, resp *http.Response) (deleteProfileResponse, error) {
	var response deleteProfileResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeGetAddressesResponse(_ context.Context, resp *http.Response) (getAddressesResponse, error) {
	var response getAddressesResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeGetAddressResponse(_ context.Context, resp *http.Response) (getAddressResponse, error) {
	var response getAddressResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodePostAddressResponse(_ context.Context, resp *http.Response) (postAddressResponse, error) {
	var response postAddressResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeDeleteAddressResponse(_ context.Context, resp *http.Response) (deleteAddressResponse, error) {
	var response deleteAddressResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encodeRequest likewise JSON-encodes the request to the HTTP request body.
// Don't use it directly as a transport/http.Client EncodeRequestFunc:
// profilesvc endpoints require mutating the HTTP method and request path.
func encodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
