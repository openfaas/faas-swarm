package handlers

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	dockerlogs "github.com/docker/cli/service/logs"
	"github.com/docker/docker/api/types"

	"github.com/openfaas/faas-provider/logs"
)

const (
	// log line prefix length frm github.com/moby/moby/pkg/stdcopy/stdcopy.go
	stdWriterPrefixLen = 8
)

// LogRequester implements the Requester interface for Swarm
type LogRequester struct {
	client ServiceLogger
}

// ServiceLogger is the subset of Docker Client methods required for querying function logs
type ServiceLogger interface {
	ServiceLogs(ctx context.Context, serviceID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
}

// NewLogRequester returns a Requestor instance that can be used in the function logs endpoint
func NewLogRequester(client ServiceLogger) logs.Requester {
	return &LogRequester{client: client}
}

// Query implements the actual Swarm logs request logic for the Requester interface
func (l LogRequester) Query(ctx context.Context, r logs.Request) (<-chan logs.Message, error) {

	options := types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: true,
		Follow:     r.Follow,
		Details:    true,
	}

	if r.Since != nil {
		options.Since = r.Since.Format(time.RFC3339)
	}

	if r.Limit > 0 {
		options.Tail = strconv.Itoa(r.Limit)
	}

	logStream, err := l.client.ServiceLogs(ctx, r.Name, options)
	if err != nil {
		return nil, err
	}

	msgStream := make(chan logs.Message)

	go parseLogStream(ctx, r.Name, msgStream, logStream)

	return msgStream, nil
}

// parseLogStream reads log lines from the logStream, parses them into Message objects, and sends
// them on the msgStream channel.  Raw log lines look like 'timestamp serviceDetails rawMessage`, e.g.
// 2019-02-09T02:34:38.914788800Z com.docker.swarm.node.id=lfplf8vfa6j2fp4xkygcze8i4,com.docker.swarm.service.id=wy8sr6u3lqx11a34t96qlbyff,com.docker.swarm.task.id=zzvbv53tdyebuhh9rquadwuud 2019/02/09 02:34:38 Error reading stdout: EOF
// we may want to pull some inspiration from here https://github.com/docker/cli/blob/master/cli/command/service/logs.go
func parseLogStream(ctx context.Context, name string, msgStream chan logs.Message, logStream io.ReadCloser) {
	defer close(msgStream)
	defer logStream.Close()

	scanner := bufio.NewScanner(logStream)
	for scanner.Scan() {
		// check if the stream was cancelled
		if ctx.Err() != nil {
			return
		}
		// trim docker log prefix

		rawMsg := string(bytes.Trim(scanner.Bytes()[stdWriterPrefixLen:], "\x00"))
		logParts := strings.SplitN(rawMsg, " ", 3)

		ts, err := time.Parse(time.RFC3339Nano, logParts[0])
		if err != nil {
			log.Printf("parseLogStream: failed to parse timestamp: %sn", err)
			return
		}

		details, err := dockerlogs.ParseLogDetails(logParts[1])
		if err != nil {
			log.Printf("parseLogStream: failed to parse log details for '%s': %s\n", rawMsg, err)
			return
		}
		msg := logs.Message{
			Name:      name, // details["com.docker.swarm.service.id"],
			Instance:  details["com.docker.swarm.task.id"],
			Timestamp: ts,
			Text:      strings.TrimSpace(logParts[2]),
		}

		msgStream <- msg
	}

	err := scanner.Err()
	if err != nil {
		log.Println("reading standard input:", err)
	}
}
