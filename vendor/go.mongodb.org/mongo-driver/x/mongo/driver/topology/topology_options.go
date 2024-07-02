// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package topology

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/internal/logger"
	"go.mongodb.org/mongo-driver/mongo/description"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
	"go.mongodb.org/mongo-driver/x/mongo/driver/ocsp"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
	"go.mongodb.org/mongo-driver/x/mongo/driver/session"
)

const defaultServerSelectionTimeout = 30 * time.Second

// Config is used to construct a topology.
type Config struct {
	Mode                   MonitorMode
	ReplicaSetName         string
	SeedList               []string
	ServerOpts             []ServerOption
	URI                    string
	ServerSelectionTimeout time.Duration
	ServerMonitor          *event.ServerMonitor
	SRVMaxHosts            int
	SRVServiceName         string
	LoadBalanced           bool
	logger                 *logger.Logger
}

// ConvertToDriverAPIOptions converts a options.ServerAPIOptions instance to a driver.ServerAPIOptions.
func ConvertToDriverAPIOptions(s *options.ServerAPIOptions) *driver.ServerAPIOptions {
	driverOpts := driver.NewServerAPIOptions(string(s.ServerAPIVersion))
	if s.Strict != nil {
		driverOpts.SetStrict(*s.Strict)
	}
	if s.DeprecationErrors != nil {
		driverOpts.SetDeprecationErrors(*s.DeprecationErrors)
	}
	return driverOpts
}

func newLogger(opts *options.LoggerOptions) (*logger.Logger, error) {
	if opts == nil {
		opts = options.Logger()
	}

	componentLevels := make(map[logger.Component]logger.Level)
	for component, level := range opts.ComponentLevels {
		componentLevels[logger.Component(component)] = logger.Level(level)
	}

	log, err := logger.New(opts.Sink, opts.MaxDocumentLength, componentLevels)
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	return log, nil
}

// NewConfig will translate data from client options into a topology config for building non-default deployments.
// Server and topology options are not honored if a custom deployment is used.
func NewConfig(co *options.ClientOptions, clock *session.ClusterClock) (*Config, error) {
	var serverAPI *driver.ServerAPIOptions

	if err := co.Validate(); err != nil {
		return nil, err
	}

	var connOpts []ConnectionOption
	var serverOpts []ServerOption

	cfgp := &Config{}

	// Set the default "ServerSelectionTimeout" to 30 seconds.
	cfgp.ServerSelectionTimeout = defaultServerSelectionTimeout

	// Set the default "SeedList" to localhost.
	cfgp.SeedList = []string{"localhost:27017"}

	// TODO(GODRIVER-814): Add tests for topology, server, and connection related options.

	// ServerAPIOptions need to be handled early as other client and server options below reference
	// c.serverAPI and serverOpts.serverAPI.
	if co.ServerAPIOptions != nil {
		serverAPI = ConvertToDriverAPIOptions(co.ServerAPIOptions)
		serverOpts = append(serverOpts, WithServerAPI(func(*driver.ServerAPIOptions) *driver.ServerAPIOptions {
			return serverAPI
		}))
	}

	cfgp.URI = co.GetURI()

	if co.SRVServiceName != nil {
		cfgp.SRVServiceName = *co.SRVServiceName
	}

	if co.SRVMaxHosts != nil {
		cfgp.SRVMaxHosts = *co.SRVMaxHosts
	}

	// AppName
	var appName string
	if co.AppName != nil {
		appName = *co.AppName

		serverOpts = append(serverOpts, WithServerAppName(func(string) string {
			return appName
		}))
	}
	// Compressors & ZlibLevel
	var comps []string
	if len(co.Compressors) > 0 {
		comps = co.Compressors

		connOpts = append(connOpts, WithCompressors(
			func(compressors []string) []string {
				return append(compressors, comps...)
			},
		))

		for _, comp := range comps {
			switch comp {
			case "zlib":
				connOpts = append(connOpts, WithZlibLevel(func(level *int) *int {
					return co.ZlibLevel
				}))
			case "zstd":
				connOpts = append(connOpts, WithZstdLevel(func(level *int) *int {
					return co.ZstdLevel
				}))
			}
		}

		serverOpts = append(serverOpts, WithCompressionOptions(
			func(opts ...string) []string { return append(opts, comps...) },
		))
	}

	var loadBalanced bool
	if co.LoadBalanced != nil {
		loadBalanced = *co.LoadBalanced
	}

	// Handshaker
	var handshaker = func(driver.Handshaker) driver.Handshaker {
		return operation.NewHello().AppName(appName).Compressors(comps).ClusterClock(clock).
			ServerAPI(serverAPI).LoadBalanced(loadBalanced)
	}
	// Auth & Database & Password & Username
	if co.Auth != nil {
		cred := &auth.Cred{
			Username:    co.Auth.Username,
			Password:    co.Auth.Password,
			PasswordSet: co.Auth.PasswordSet,
			Props:       co.Auth.AuthMechanismProperties,
			Source:      co.Auth.AuthSource,
		}
		mechanism := co.Auth.AuthMechanism

		if len(cred.Source) == 0 {
			switch strings.ToUpper(mechanism) {
			case auth.MongoDBX509, auth.GSSAPI, auth.PLAIN:
				cred.Source = "$external"
			default:
				cred.Source = "admin"
			}
		}

		authenticator, err := auth.CreateAuthenticator(mechanism, cred)
		if err != nil {
			return nil, err
		}

		handshakeOpts := &auth.HandshakeOptions{
			AppName:       appName,
			Authenticator: authenticator,
			Compressors:   comps,
			ServerAPI:     serverAPI,
			LoadBalanced:  loadBalanced,
			ClusterClock:  clock,
			HTTPClient:    co.HTTPClient,
		}

		if mechanism == "" {
			// Required for SASL mechanism negotiation during handshake
			handshakeOpts.DBUser = cred.Source + "." + cred.Username
		}
		if co.AuthenticateToAnything != nil && *co.AuthenticateToAnything {
			// Authenticate arbiters
			handshakeOpts.PerformAuthentication = func(serv description.Server) bool {
				return true
			}
		}

		handshaker = func(driver.Handshaker) driver.Handshaker {
			return auth.Handshaker(nil, handshakeOpts)
		}
	}
	connOpts = append(connOpts, WithHandshaker(handshaker))
	// ConnectTimeout
	if co.ConnectTimeout != nil {
		serverOpts = append(serverOpts, WithHeartbeatTimeout(
			func(time.Duration) time.Duration { return *co.ConnectTimeout },
		))
		connOpts = append(connOpts, WithConnectTimeout(
			func(time.Duration) time.Duration { return *co.ConnectTimeout },
		))
	}
	// Dialer
	if co.Dialer != nil {
		connOpts = append(connOpts, WithDialer(
			func(Dialer) Dialer { return co.Dialer },
		))
	}
	// Direct
	if co.Direct != nil && *co.Direct {
		cfgp.Mode = SingleMode
	}

	// HeartbeatInterval
	if co.HeartbeatInterval != nil {
		serverOpts = append(serverOpts, WithHeartbeatInterval(
			func(time.Duration) time.Duration { return *co.HeartbeatInterval },
		))
	}
	// Hosts
	cfgp.SeedList = []string{"localhost:27017"} // default host
	if len(co.Hosts) > 0 {
		cfgp.SeedList = co.Hosts
	}

	// MaxConIdleTime
	if co.MaxConnIdleTime != nil {
		serverOpts = append(serverOpts, WithConnectionPoolMaxIdleTime(
			func(time.Duration) time.Duration { return *co.MaxConnIdleTime },
		))
	}
	// MaxPoolSize
	if co.MaxPoolSize != nil {
		serverOpts = append(
			serverOpts,
			WithMaxConnections(func(uint64) uint64 { return *co.MaxPoolSize }),
		)
	}
	// MinPoolSize
	if co.MinPoolSize != nil {
		serverOpts = append(
			serverOpts,
			WithMinConnections(func(uint64) uint64 { return *co.MinPoolSize }),
		)
	}
	// MaxConnecting
	if co.MaxConnecting != nil {
		serverOpts = append(
			serverOpts,
			WithMaxConnecting(func(uint64) uint64 { return *co.MaxConnecting }),
		)
	}
	// PoolMonitor
	if co.PoolMonitor != nil {
		serverOpts = append(
			serverOpts,
			WithConnectionPoolMonitor(func(*event.PoolMonitor) *event.PoolMonitor { return co.PoolMonitor }),
		)
	}
	// Monitor
	if co.Monitor != nil {
		connOpts = append(connOpts, WithMonitor(
			func(*event.CommandMonitor) *event.CommandMonitor { return co.Monitor },
		))
	}
	// ServerMonitor
	if co.ServerMonitor != nil {
		serverOpts = append(
			serverOpts,
			WithServerMonitor(func(*event.ServerMonitor) *event.ServerMonitor { return co.ServerMonitor }),
		)
		cfgp.ServerMonitor = co.ServerMonitor
	}
	// ReplicaSet
	if co.ReplicaSet != nil {
		cfgp.ReplicaSetName = *co.ReplicaSet
	}
	// ServerSelectionTimeout
	if co.ServerSelectionTimeout != nil {
		cfgp.ServerSelectionTimeout = *co.ServerSelectionTimeout
	}
	// SocketTimeout
	if co.SocketTimeout != nil {
		connOpts = append(
			connOpts,
			WithReadTimeout(func(time.Duration) time.Duration { return *co.SocketTimeout }),
			WithWriteTimeout(func(time.Duration) time.Duration { return *co.SocketTimeout }),
		)
	}
	// TLSConfig
	if co.TLSConfig != nil {
		connOpts = append(connOpts, WithTLSConfig(
			func(*tls.Config) *tls.Config {
				return co.TLSConfig
			},
		))
	}

	// HTTP Client
	if co.HTTPClient != nil {
		connOpts = append(connOpts, WithHTTPClient(
			func(*http.Client) *http.Client {
				return co.HTTPClient
			},
		))
	}

	// OCSP cache
	ocspCache := ocsp.NewCache()
	connOpts = append(
		connOpts,
		WithOCSPCache(func(ocsp.Cache) ocsp.Cache { return ocspCache }),
	)

	// Disable communication with external OCSP responders.
	if co.DisableOCSPEndpointCheck != nil {
		connOpts = append(
			connOpts,
			WithDisableOCSPEndpointCheck(func(bool) bool { return *co.DisableOCSPEndpointCheck }),
		)
	}

	// LoadBalanced
	if co.LoadBalanced != nil {
		cfgp.LoadBalanced = *co.LoadBalanced

		serverOpts = append(
			serverOpts,
			WithServerLoadBalanced(func(bool) bool { return *co.LoadBalanced }),
		)
		connOpts = append(
			connOpts,
			WithConnectionLoadBalanced(func(bool) bool { return *co.LoadBalanced }),
		)
	}

	lgr, err := newLogger(co.LoggerOptions)
	if err != nil {
		return nil, err
	}

	serverOpts = append(
		serverOpts,
		withLogger(func() *logger.Logger { return lgr }),
		withServerMonitoringMode(co.ServerMonitoringMode),
	)

	cfgp.logger = lgr

	serverOpts = append(
		serverOpts,
		WithClock(func(*session.ClusterClock) *session.ClusterClock { return clock }),
		WithConnectionOptions(func(...ConnectionOption) []ConnectionOption { return connOpts }))

	cfgp.ServerOpts = serverOpts

	return cfgp, nil
}
