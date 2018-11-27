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
	// lister is ServiceLister client, typically the Docker client
	lister ServiceLister
	// dnsRoundRobin controls  if DNSRoundRobin is used to resolve function
	dnsRoundRobin bool
	// scheme is the http scheme (http/https) used to proxy the request
	scheme string
	// dnsrrLookup method used to resolve the function IP address, defaults to the internal lookupIP
	// method, which is an implementation of net.LookupIP
	dnsrrLookup func(context.Context, string) ([]net.IP, error)
}

// NewFunctionLookup creates a new FunctionLookup resolver
func NewFunctionLookup(client ServiceLister, dnsRoundRobin bool) *FunctionLookup {
	return &FunctionLookup{
		lister:        client,
		dnsRoundRobin: dnsRoundRobin,
		scheme:        urlScheme,
		dnsrrLookup:   lookupIP,
	}
}

// Resolve implements the openfaas-provider proxy.BaseURLResolver interface. In
// short it verifies that a function with the given name is resolvable by Docker
// Swarm.  It can be configured to do this via DNS or by querying the Docker Service
// list.
func (l *FunctionLookup) Resolve(name string) (u url.URL, err error) {
	return l.ResolveContext(context.Background(), name)
}

// ResolveContext provides an implementation of openfaas-provider proxy.BaseURLResolver with
// context support. See `Resolve`
func (l *FunctionLookup) ResolveContext(ctx context.Context, name string) (u url.URL, err error) {

	if l.dnsRoundRobin {
		u.Host, err = l.byDNSRoundRobin(ctx, name)
	} else {
		u.Host, err = l.byName(ctx, name)
	}

	if err != nil {
		return u, err
	}

	u.Scheme = l.scheme
	return u, nil
}

// resolve the function by checking the available docker VIP based resolution
func (l *FunctionLookup) byName(ctx context.Context, name string) (string, error) {
	serviceFilter := filters.NewArgs()
	serviceFilter.Add("name", name)
	services, err := l.lister.ServiceList(ctx, types.ServiceListOptions{Filters: serviceFilter})

	if err != nil {
		return "", err
	}

	if len(services) > 0 {
		return name, nil
	}

	return "", fmt.Errorf("could not resolve: %s", name)
}

// resolve the function by checking the available docker DNSRR resolution
func (l *FunctionLookup) byDNSRoundRobin(ctx context.Context, name string) (string, error) {
	entries, lookupErr := l.dnsrrLookup(ctx, fmt.Sprintf("tasks.%s", name))

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

// lookupIP implements the net.LookupIP method with context support. It returns a slice of that\
// host's IPv4 and IPv6 addresses.
func lookupIP(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, ia := range addrs {
		ips[i] = ia.IP
	}
	return ips, nil
}
