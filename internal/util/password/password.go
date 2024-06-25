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

// Package password provides utilities for password hashing and verification.
package password

// Password wraps a password string.
//
// It exist mainly to avoid issues when multiple string parameters are used.
//
// It should be passed by value.
type Password struct {
	p string
}

// WrapPassword returns Password for the given string.
func WrapPassword(password string) Password {
	return Password{p: password}
}

// Password returns the password string.
func (p Password) Password() string {
	return p.p
}
