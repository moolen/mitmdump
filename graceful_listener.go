package mitmdump

import (
	"net"
	"sync"
)

// GracefulListener serves a stoppable connection and tracks their lifetime to notify
// when it is safe to terminate the application.
type GracefulListener struct {
	net.Listener
	sync.WaitGroup
}

type stoppableConn struct {
	net.Conn
	wg *sync.WaitGroup
}

// NewGracefulListener is a factory function for a graceful listener
func NewGracefulListener(l net.Listener) *GracefulListener {
	return &GracefulListener{l, sync.WaitGroup{}}
}

// Accept overrides the net.Conn.Accept() interface
func (sl *GracefulListener) Accept() (net.Conn, error) {
	c, err := sl.Listener.Accept()
	if err != nil {
		return c, err
	}
	sl.Add(1)
	return &stoppableConn{c, &sl.WaitGroup}, nil
}

func (sc *stoppableConn) Close() error {
	sc.wg.Done()
	return sc.Conn.Close()
}
