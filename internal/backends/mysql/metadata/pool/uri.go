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

package pool

import (
	"github.com/go-sql-driver/mysql"
	"net/url"
	"path"
)

func parseURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	username := u.User.Username()
	password, _ := u.User.Password()

	dbName := path.Clean(u.Path)

	values := u.Query()
	params := make(map[string]string, len(values))

	for k := range values {
		params[k] = values.Get(k)
	}

	// mysql url requires a specified format to work
	// For example: username:password@tcp(127.0.0.1:3306)/ferretdb
	cfg := mysql.Config{
		User:   username,
		Passwd: password,
		Net:    "tcp",
		Addr:   u.Host,
		DBName: dbName,
		Params: params,
	}
	mysqlURL := cfg.FormatDSN()

	return mysqlURL, nil
}
