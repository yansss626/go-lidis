package lidis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/silenceper/pool"
)

const (
	defaultInitialCap  = 0
	defaultMaxIdle     = 5
	defaultMaxCap      = 20
	defaultIdleTimeout = 10 * time.Minute
)

type ClientPool struct {
	pool ConnectionPool
}

type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
	closed bool
}

type ConnectionPool interface {
	Get() (interface{}, error)
	Put(interface{}) error
	Close(interface{}) error
	Release()
}

type ConnectionPoolConfig struct {
	Host        string
	Port        int
	InitialCap  int
	MaxIdle     int
	MaxCap      int
	IdleTimeout time.Duration
}

func newConnectionPool(cnf *ConnectionPoolConfig) (ConnectionPool, error) {

	factory := func() (interface{}, error) {
		addr := net.JoinHostPort(cnf.Host, strconv.Itoa(cnf.Port))
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return &Connection{
			conn:   conn,
			reader: bufio.NewReader(conn),
		}, nil
	}

	close := func(v interface{}) error {

		client, ok := v.(*Connection)
		if !ok {
			return fmt.Errorf("invalid connection type")
		}

		return client.conn.Close()

	}

	poolConfig := &pool.Config{
		InitialCap:  cnf.InitialCap,
		MaxIdle:     cnf.MaxIdle,
		MaxCap:      cnf.MaxCap,
		Factory:     factory,
		Close:       close,
		IdleTimeout: cnf.IdleTimeout,
	}
	pool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func NewClientPool(cnf *ConnectionPoolConfig) (*ClientPool, error) {

	config := *cnf
	if cnf.IdleTimeout == 0 {
		config.IdleTimeout = defaultIdleTimeout
	}
	if cnf.InitialCap == 0 {
		config.InitialCap = defaultInitialCap
	}
	if cnf.MaxCap == 0 {
		config.MaxCap = defaultMaxCap
	}
	if cnf.MaxIdle == 0 {
		config.MaxIdle = defaultMaxIdle
	}

	pool, err := newConnectionPool(&config)
	if err != nil {
		return nil, err
	}
	return &ClientPool{
		pool: pool,
	}, nil
}

func NewConnection(addr string) (*Connection, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (p *ClientPool) Get() (*Connection, func(), error) {
	for {
		obj, err := p.pool.Get()
		if err != nil {
			return nil, nil, err
		}
		client, ok := obj.(*Connection)
		if !ok {
			p.pool.Close(obj)
			return nil, nil, fmt.Errorf("failed to type assertion Connection")
		}
		if client.closed {
			p.pool.Close(obj)
			continue
		}
		released := false
		return client, func() {
			if released {
				return
			}
			released = true
			if client.closed {
				p.discard(client)
			} else {
				p.put(client)
			}
		}, nil
	}

}

// 放回一个连接
func (p *ClientPool) put(client *Connection) error {
	return p.pool.Put(client)
}

// 关闭一个连接
func (p *ClientPool) discard(client *Connection) error {
	return p.pool.Close(client)
}

// 释放连接池
func (p *ClientPool) Release() {
	p.pool.Release()
}

func (c *Connection) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

func (c *Connection) markError(err error) {
	if isNetWorkError(err) {
		c.closed = true
	}
}

func isNetWorkError(err error) bool {
	return errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF)
}
