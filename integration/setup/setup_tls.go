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
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func generateTLSPair() ([]byte, []byte) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"FerretDB"},
		},
		NotBefore: time.Now(),
		// Make it valid for short amount of time to avoid accidental use.
		NotAfter: time.Now().Add(time.Minute * 20),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, key.PublicKey, key)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	certBytes := &bytes.Buffer{}

	err = pem.Encode(certBytes, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		panic(err)
	}

	privateKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic(err)
	}

	privateKeyBytes := &bytes.Buffer{}

	err = pem.Encode(privateKeyBytes, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKey})
	if err != nil {
		panic(err)
	}

	return certBytes.Bytes(), privateKeyBytes.Bytes()
}

// GetTLSFilesPaths returns paths to TLS files.
func GetTLSFilesPaths() (string, string) {
	cert, key := generateTLSPair()

	var certPath, keyPath = "cert.pem", "key.pem"

	err := os.WriteFile(certPath, cert, 0644)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(keyPath, key, 0644)
	if err != nil {
		panic(err)
	}

	return certPath, keyPath
}
