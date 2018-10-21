package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
)

const watchdogPort = 8080
const urlScheme = "http"

// ServiceLister is the subset of the Docker client.ServiceAPIClient needed to enable the
// function lookup
type ServiceLister interface {
	ServiceList(context.Context, types.ServiceListOptions) ([]swarm.Service, error)
}

// FunctionLookup is a openfaas-provider proxy.BaseURLResolver that allows the
// caller to verify that a function is resolvable.
type FunctionLookup struct {
	docker      ServiceLister
	dnsrr       bool
	scheme      string
	dnsrrLookup func(string) ([]net.IP, error)
}

// NewFunctionLookup instantiates a new FunctionLookup resolver
func NewFunctionLookup(client ServiceLister, dnsrr bool) *FunctionLookup {
	return &FunctionLookup{
		docker:      client,
		dnsrr:       dnsrr,
		scheme:      urlScheme,
		dnsrrLookup: net.LookupIP,
	}
}

// Resolve implements the openfaas-provider proxy.BaseURLResolver interface. In
// short it verifies that a function with the given name is resolvable by Docker
// Swarm.  It can be configured to do this via DNS or by querying the Docker Service
// list.
func (l *FunctionLookup) Resolve(name string) (u url.URL, err error) {

	if l.dnsrr {
		u.Host, err = l.byDNSRoundRobin(name)
	} else {
		u.Host, err = l.byName(name)
	}

	if err != nil {
		return u, err
	}

	u.Scheme = l.scheme
	return u, nil
}

// resolve the function by checking the available docker VIP based resolution
func (l *FunctionLookup) byName(name string) (string, error) {
	serviceFilter := filters.NewArgs()
	serviceFilter.Add("name", name)
	services, err := l.docker.ServiceList(context.Background(), types.ServiceListOptions{Filters: serviceFilter})

	if err != nil {
		return "", err
	}

	if len(services) > 0 {
		return name, nil
	}

	return "", fmt.Errorf("could not resolve: %s", name)
}

// resolve the function by checking the available docker DNSRR resolution
func (l *FunctionLookup) byDNSRoundRobin(name string) (string, error) {
	entries, lookupErr := l.dnsrrLookup(fmt.Sprintf("tasks.%s", name))

	if lookupErr != nil {
		return "", lookupErr
	}

	if len(entries) > 0 {
		index := randomInt(0, len(entries))
		return entries[index].String(), nil
	}

	return "", fmt.Errorf("could not resolve '%s' using dnsrr", name)
}

func randomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
