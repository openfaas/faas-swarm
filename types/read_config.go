// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package types

import (
	"os"
	"strconv"
	"time"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

// HasEnv provides interface for os.Getenv
type HasEnv interface {
	Getenv(key string) string
}

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

func parseIntValue(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}

	return duration
}

func parseBoolValue(val string, fallback bool) bool {
	switch val {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	case "0", "f", "F", "false", "FALSE", "False":
		return false
	default:
		return fallback
	}
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) BootstrapConfig {
	cfg := BootstrapConfig{}

	readTimeout := parseIntOrDurationValue(hasEnv.Getenv("read_timeout"), time.Second*10)
	writeTimeout := parseIntOrDurationValue(hasEnv.Getenv("write_timeout"), time.Second*10)

	const defaultPort = 8080

	cfg.TCPPort = parseIntValue(hasEnv.Getenv("port"), defaultPort)
	cfg.ReadTimeout = readTimeout
	cfg.WriteTimeout = writeTimeout

	cfg.EnableBasicAuth = parseBoolValue(hasEnv.Getenv("basic_auth"), false)
	cfg.DNSrr = parseBoolValue(hasEnv.Getenv("dnsrr"), false)

	return cfg
}

// BootstrapConfig for the process.
type BootstrapConfig struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	TCPPort         int
	EnableBasicAuth bool
	DNSrr           bool
}
