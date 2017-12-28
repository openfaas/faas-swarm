package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"

	"github.com/openfaas/faas/gateway/requests"
)

const watchdogPort = 8080

//FunctionProxy passes-through to functions
func FunctionProxy(wildcard bool, client *client.Client) http.HandlerFunc {

	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          1,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case "POST", "GET":
			log.Print(r.Header)

			xFunctionHeader := r.Header["X-Function"]
			if len(xFunctionHeader) > 0 {
				log.Print("X-Function: ", xFunctionHeader)
			}

			vars2 := mux.Vars(r)
			// service := vars2["name"]
			fmt.Println(vars2)
			// getServiceName
			var serviceName string
			if wildcard {
				vars := mux.Vars(r)
				fmt.Println("vars ", vars)
				name := vars["name"]
				serviceName = name
			} else if len(xFunctionHeader) > 0 {
				serviceName = xFunctionHeader[0]
			}

			if len(serviceName) > 0 {
				lookupInvoke(w, r, serviceName, client, &proxyClient)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Provide an x-function header or valid route /function/function_name."))
			}
			break
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func lookupInvoke(w http.ResponseWriter, r *http.Request, name string, c *client.Client, proxyClient *http.Client) {
	exists, err := lookupSwarmService(name, c)

	if err != nil || exists == false {
		if err != nil {
			log.Printf("Could not resolve service: %s error: %s.", name, err)
		}

		// TODO: Should record the 404/not found error in Prometheus.
		writeHead(name, http.StatusNotFound, w)
		w.Write([]byte(fmt.Sprintf("Cannot find service: %s.", name)))
	}

	if exists {
		forwardReq := requests.NewForwardRequest(r.Method, *r.URL)

		invokeService(w, r, name, forwardReq, proxyClient)
	}
}

func lookupSwarmService(serviceName string, c *client.Client) (bool, error) {
	fmt.Printf("Resolving: '%s'\n", serviceName)
	serviceFilter := filters.NewArgs()
	serviceFilter.Add("name", serviceName)
	services, err := c.ServiceList(context.Background(), types.ServiceListOptions{Filters: serviceFilter})

	return len(services) > 0, err
}

func invokeService(w http.ResponseWriter, r *http.Request, service string, forwardReq requests.ForwardRequest, proxyClient *http.Client) {
	stamp := strconv.FormatInt(time.Now().Unix(), 10)

	//TODO: inject setting rather than looking up each time.
	var dnsrr bool
	if os.Getenv("dnsrr") == "true" {
		dnsrr = true
	}

	addr := service
	// Use DNS-RR via tasks.servicename if enabled as override, otherwise VIP.
	if dnsrr {
		entries, lookupErr := net.LookupIP(fmt.Sprintf("tasks.%s", service))
		if lookupErr == nil && len(entries) > 0 {
			index := randomInt(0, len(entries))
			addr = entries[index].String()
		}
	}

	url := forwardReq.ToURL(addr, watchdogPort)

	contentType := r.Header.Get("Content-Type")
	fmt.Printf("[%s] Forwarding request [%s] to: %s\n", stamp, contentType, url)

	if r.Body != nil {
		defer r.Body.Close()
	}

	request, err := http.NewRequest(r.Method, url, r.Body)

	copyHeaders(&request.Header, &r.Header)

	response, err := proxyClient.Do(request)
	if err != nil {
		log.Print(err)
		writeHead(service, http.StatusInternalServerError, w)
		buf := bytes.NewBufferString("Can't reach service: " + service)
		w.Write(buf.Bytes())
		return
	}

	clientHeader := w.Header()
	copyHeaders(&clientHeader, &response.Header)

	defaultHeader := "text/plain"

	w.Header().Set("Content-Type", GetContentType(response.Header, r.Header, defaultHeader))

	writeHead(service, response.StatusCode, w)

	if response.Body != nil {
		io.Copy(w, response.Body)
	}
}

// GetContentType resolves the correct Content-Tyoe for a proxied function
func GetContentType(request http.Header, proxyResponse http.Header, defaultValue string) string {
	responseHeader := proxyResponse.Get("Content-Type")
	requestHeader := request.Get("Content-Type")

	var headerContentType string
	if len(responseHeader) > 0 {
		headerContentType = responseHeader
	} else if len(requestHeader) > 0 {
		headerContentType = requestHeader
	} else {
		headerContentType = defaultValue
	}

	return headerContentType
}

func copyHeaders(destination *http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(*destination)[k] = vClone
	}
}

func randomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func writeHead(service string, code int, w http.ResponseWriter) {
	w.WriteHeader(code)
}
