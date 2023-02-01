// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package setup

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// TLSFileSide defines a "side" of TLS file (client or server).
type TLSFileSide string

// Possible values for TLSFileSide.
const (
	ClientSide TLSFileSide = "client"
	ServerSide TLSFileSide = "server"
)

// TLSFilesPaths returns paths to TLS files - cert, key, ca.
type TLSFilesPaths struct {
	Cert string
	Key  string
	CA   string
}

// GetTLSFilesPaths returns paths to TLS files - cert, key, ca.
func GetTLSFilesPaths(side TLSFileSide) (*TLSFilesPaths, error) {
	certPath := filepath.Join("..", "build", "certs", string(side)+"-cert.pem")

	if _, err := os.Stat(certPath); err != nil {
		return nil, lazyerrors.Error(err)
	}

	keyPath := filepath.Join("..", "build", "certs", string(side)+"-key.pem")

	if _, err := os.Stat(keyPath); err != nil {
		return nil, lazyerrors.Error(err)
	}

	caPath := filepath.Join("..", "build", "certs", "rootCA.pem")

	if _, err := os.Stat(keyPath); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &TLSFilesPaths{
		Cert: certPath,
		Key:  keyPath,
		CA:   caPath,
	}, nil
}

// GetClientTLSConfig returns a test TLS config for a client.
func GetClientTLSConfig() (*tls.Config, error) {
	tlsFiles, err := GetTLSFilesPaths(ClientSide)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Load the root CA certificate
	rootCA, err := os.ReadFile(tlsFiles.CA)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootCA)
	if !ok {
		return nil, lazyerrors.New("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(tlsFiles.Cert, tlsFiles.Key)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Create the TLS config
	tlsConfig := &tls.Config{
		RootCAs:      roots,
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig, nil
}
