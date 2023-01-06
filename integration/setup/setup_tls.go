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
	"testing"

	"github.com/stretchr/testify/require"
)

// TLSFileSide defines a "side" of TLS file (client or server).
type TLSFileSide string

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
func GetTLSFilesPaths(t testing.TB, side TLSFileSide) *TLSFilesPaths {
	certPath := filepath.Join("..", "build", "certs", string(side)+"-cert.pem")

	_, err := os.Stat(certPath)
	require.NoError(t, err)

	keyPath := filepath.Join("..", "build", "certs", string(side)+"-key.pem")

	_, err = os.Stat(keyPath)
	require.NoError(t, err)

	caPath := filepath.Join("..", "build", "certs", "rootCA.pem")

	_, err = os.Stat(keyPath)
	require.NoError(t, err)

	return &TLSFilesPaths{
		Cert: certPath,
		Key:  keyPath,
		CA:   caPath,
	}
}

func GetClientTLSConfig(t testing.TB) *tls.Config {
	tlsFiles := GetTLSFilesPaths(t, ClientSide)

	// Load the root CA certificate
	rootCA, err := os.ReadFile(tlsFiles.CA)
	require.NoError(t, err)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootCA)
	require.True(t, ok, "failed to parse root certificate")

	cert, err := tls.LoadX509KeyPair(tlsFiles.Cert, tlsFiles.Key)
	require.NoError(t, err)

	// Create the TLS config
	tlsConfig := &tls.Config{
		RootCAs:      roots,
		Certificates: []tls.Certificate{cert},
		// InsecureSkipVerify: true,
	}

	return tlsConfig
}
