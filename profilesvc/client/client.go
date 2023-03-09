// Package client provides a profilesvc client based on a predefined Consul
// service name and relevant tags. Users must only provide the address of a
// Consul server.
package client

import (
	"io"
	"time"

	tendpoint "github.com/RangelReale/go-kit-typed/endpoint"
	consulapi "github.com/hashicorp/consul/api"

	"github.com/go-kit/examples/profilesvc"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

// New returns a service that's load-balanced over instances of profilesvc found
// in the provided Consul server. The mechanism of looking up profilesvc
// instances in Consul is hard-coded into the client.
func New(consulAddr string, logger log.Logger) (profilesvc.Service, error) {
	apiclient, err := consulapi.NewClient(&consulapi.Config{
		Address: consulAddr,
	})
	if err != nil {
		return nil, err
	}

	// As the implementer of profilesvc, we declare and enforce these
	// parameters for all of the profilesvc consumers.
	var (
		consulService = "profilesvc"
		consulTags    = []string{"prod"}
		passingOnly   = true
		retryMax      = 3
		retryTimeout  = 500 * time.Millisecond
	)

	var (
		sdclient  = consul.NewClient(apiclient)
		instancer = consul.NewInstancer(sdclient, logger, consulService, consulTags, passingOnly)
		endpoints profilesvc.Endpoints
	)
	{
		factory := factoryFor(profilesvc.MakePostProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PostProfileEndpoint = tendpoint.Cast(endpoints.PostProfileEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakeGetProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetProfileEndpoint = tendpoint.Cast(endpoints.GetProfileEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakePutProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PutProfileEndpoint = tendpoint.Cast(endpoints.PutProfileEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakePatchProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PatchProfileEndpoint = tendpoint.Cast(endpoints.PatchProfileEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakeDeleteProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.DeleteProfileEndpoint = tendpoint.Cast(endpoints.DeleteProfileEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakeGetAddressesEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetAddressesEndpoint = tendpoint.Cast(endpoints.GetAddressesEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakeGetAddressEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetAddressEndpoint = tendpoint.Cast(endpoints.GetAddressEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakePostAddressEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PostAddressEndpoint = tendpoint.Cast(endpoints.PostAddressEndpoint, retry)
	}
	{
		factory := factoryFor(profilesvc.MakeDeleteAddressEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.DeleteAddressEndpoint = tendpoint.Cast(endpoints.DeleteAddressEndpoint, retry)
	}

	return endpoints, nil
}

func factoryFor[Req any, Resp any](makeEndpoint func(profilesvc.Service) tendpoint.Endpoint[Req, Resp]) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		service, err := profilesvc.MakeClientEndpoints(instance)
		if err != nil {
			return nil, nil, err
		}
		return tendpoint.ReverseAdapter(makeEndpoint(service)), nil, nil
	}
}
