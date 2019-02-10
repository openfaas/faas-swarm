package logs

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/openfaas/faas-provider/httputils"
)

// Requestor submits queries the logging system.
// This will be passed to the log handler constructor.
type Requestor interface {
	// Filter allows the log handler to provide additional server side filtering of Messages.
	Filter(Request, Message) bool
	// Query submits a log request to the actual logging system.
	Query(context.Context, Request) (<-chan Message, error)
}

// NewLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
func NewLogHandlerFunc(requestor Requestor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			log.Println("LogHandler: response is not a CloseNotifier, required for streaming response")
			http.NotFound(w, r)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Println("LogHandler: response is not a Flusher, required for streaming response")
			http.NotFound(w, r)
			return
		}

		logRequest := Request{}
		err := json.NewDecoder(r.Body).Decode(&logRequest)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			httputils.WriteError(w, http.StatusUnprocessableEntity, "could not parse the log request")
			return
		}

		// magic here
		ctx, cancelQuery := context.WithCancel(r.Context())
		defer cancelQuery()
		messages, err := requestor.Query(ctx, logRequest)
		if err != nil {
			// add smarter error handling here
			httputils.WriteError(w, http.StatusInternalServerError, "function log request failed")
			return
		}

		// Send the initial headers saying we're gonna stream the response.
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		sent := 0
		jsonEncoder := json.NewEncoder(w)

		for messages != nil {
			select {
			case <-cn.CloseNotify():
				log.Println("LogHandler: client stopped listening")
				return
			case msg, ok := <-messages:
				if !ok {
					log.Println("LogHandler: end of log stream")
					messages = nil
					return
				}
				// maybe skip the filtering here and require the Query method to handle all of the filtering?
				if !requestor.Filter(logRequest, msg) {
					continue
				}
				// serialize and write the msg to the http ResponseWriter
				err := jsonEncoder.Encode(msg)
				if err != nil {
					// can't actually write the status header here so we should json serialize an error
					// and return that because we have already sent the content type and status code
					log.Printf("LogHandler: failed to serialize log message: '%s'\n", msg.String())
					// write json error message here ?
					jsonEncoder.Encode(Message{Text: "failed to serialize log message"})
					return
				}

				flusher.Flush()

				sent++
				if logRequest.Limit > 0 && sent >= logRequest.Limit {
					return
				}
			}
		}
	}
}
