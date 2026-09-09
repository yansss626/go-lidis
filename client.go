package lidis

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/silenceper/pool"
)

type Client struct {
	pool ConnectionPool
}

type ConnectionPool interface {
	Get() (interface{}, error)
	Put(interface{}) error
	Close(interface{}) error
}

type connection struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

type ConnectionPoolConfig struct {
	Host        string
	Port        int
	InitialCap  int
	MaxIdle     int
	MaxCap      int
	IdleTimeout int64
}

func NewConnectionPool(cnf *ConnectionPoolConfig) (ConnectionPool, error) {

	factory := func() (interface{}, error) {
		addr := net.JoinHostPort(cnf.Host, strconv.Itoa(cnf.Port))
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return &connection{
			Conn:   conn,
			Reader: bufio.NewReader(conn),
		}, nil
	}

	close := func(v interface{}) error {

		conn, ok := v.(*connection)
		if !ok {
			return fmt.Errorf("invalid connection type")
		}

		return conn.Conn.Close()

	}

	poolConfig := &pool.Config{
		InitialCap:  cnf.InitialCap,
		MaxIdle:     cnf.MaxIdle,
		MaxCap:      cnf.MaxCap,
		Factory:     factory,
		Close:       close,
		IdleTimeout: time.Duration(cnf.IdleTimeout) * time.Minute,
	}
	pool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func NewClient(pool ConnectionPool) *Client {
	return &Client{
		pool: pool,
	}
}

func (c *Client) acquireConnection() (*connection, func(), error) {
	obj, err := c.pool.Get()
	if err != nil {
		return nil, nil, err
	}
	conn, ok := obj.(*connection)
	if !ok {
		c.pool.Close(obj)
		return nil, nil, fmt.Errorf("invalid connection type")
	}

	release := func() {
		c.pool.Put(obj)
	}

	return conn, release, nil
}
