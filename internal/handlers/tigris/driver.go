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

package tigris

import (
	"context"

	"github.com/tigrisdata/tigrisdb-client-go/driver"
	"go.uber.org/zap"
)

type Client struct {
	conn driver.Driver
}

// NewConn creates a gRPC connection to the Tigris database backend.
func NewConn(connString string, logger *zap.Logger, lazy bool) (*Client, error) {
	ctx := context.TODO()
	conf := new(driver.Config)
	c, err := driver.NewDriver(ctx, connString, conf)
	if err != nil {
		panic(err)
	}
	res := &Client{
		conn: c,
	}
	if !lazy {
		err = res.Check(ctx)
	}
	return res, err
}

func (cli *Client) Close() error {
	return cli.conn.Close()
}

func (cli *Client) Check(ctx context.Context) error {
	_, err := cli.conn.ListDatabases(ctx)
	return err
}
