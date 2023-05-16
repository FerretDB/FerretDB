package sdb

import (
	"database/sql"
	"errors"
	"sync"
)

func newConnPool() *connPool {
	return &connPool{
		mx:  sync.Mutex{},
		dbs: make(map[string]*sql.DB),
	}
}

type connPool struct {
	mx  sync.Mutex
	dbs map[string]*sql.DB
}

func (c *connPool) DB(name string) (*sql.DB, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if db, ok := c.dbs[name]; ok {
		return db, nil
	}

	db, err := sql.Open("sqlite", name)
	if err != nil {
		return nil, err
	}

	c.dbs[name] = db

	return db, nil
}

func (c *connPool) Close() error {
	var errs error

	c.mx.Lock()
	defer c.mx.Unlock()

	for _, conn := range c.dbs {
		if err := conn.Close(); err != nil {
			errors.Join(err)
		}
	}

	return errs
}
