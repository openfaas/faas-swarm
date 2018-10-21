package handlers

import (
	"context"
	"errors"
	"net"
	"testing"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

type mockServiceLister struct {
	serviceName string
	resolverErr bool
	err         error
}

func (l mockServiceLister) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	if l.resolverErr {
		return nil, l.err
	}

	if options.Filters.Get("name")[0] != l.serviceName {
		return nil, nil
	}

	return []swarm.Service{{ID: l.serviceName}}, nil
}

func Test_ProxyURLResolver_ByName(t *testing.T) {

	scenarios := []struct {
		name        string
		fncName     string
		resolverErr bool
		err         error
	}{
		{"returns url with function name as host", "testFnc", false, nil},
		{"returns error if function is not resolved", "sampleFnc", false, errors.New("could not resolve: testFnc")},
		{"returns error if docker lookup fails", "", true, errors.New("docker search error")},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			docker := mockServiceLister{s.fncName, s.resolverErr, s.err}
			u, err := NewFunctionLookup(docker, false).Resolve("testFnc")
			if err != nil && err.Error() != s.err.Error() {
				t.Errorf("expected resolver error `%s`, got `%s`", s.err, err)
			}

			if s.err == nil && u.Host != s.fncName {
				t.Errorf("expected url host `%s`, got `%s`", s.fncName, u.Host)
			}
		})
	}
}

func Test_ProxyURLResolver_RoundRobingErrs(t *testing.T) {

	scenarios := []struct {
		name        string
		fncName     string
		ipAddr      string
		resolverErr bool
		err         error
	}{
		{"returns ip address", "dnsrrTestFncExists", "0.0.0.0", false, nil},
		{"returns error if lookup errors", "dnsrrTestFncDoesNotExist", "", true, errors.New("lookup tasks.dnsrrTestFncDoesNotExist: no such host")},
		{"returns error if no ips returned", "dnsrrTestFncNoIP", "", true, errors.New("could not resolve 'dnsrrTestFncNoIP' using dnsrr")},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			docker := mockServiceLister{s.fncName, s.resolverErr, s.err}
			resolver := NewFunctionLookup(docker, true)
			resolver.dnsrrLookup = testDNSRRLookup

			u, err := resolver.Resolve(s.fncName)
			if err != nil && err.Error() != s.err.Error() {
				t.Errorf("expected resolver error `%s`, got `%s`", s.err, err)
			}

			if s.err == nil && u.Host != s.ipAddr {
				t.Errorf("expected url host `%s`, got `%s`", s.ipAddr, u.Host)
			}
		})
	}
}

func testDNSRRLookup(name string) ([]net.IP, error) {
	if name == "tasks.dnsrrTestFncDoesNotExist" {
		return nil, errors.New("lookup tasks.dnsrrTestFncDoesNotExist: no such host")
	}

	if name == "tasks.dnsrrTestFncExists" {
		return []net.IP{net.IPv4zero}, nil
	}

	return []net.IP{}, nil
}
