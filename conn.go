package connPool

import (
	"net"
	"sync/atomic"
	"time"
)

var noDeadline = time.Time{}

type Conn struct {
	netConn net.Conn
	usedAt  atomic.Value
}

func NewConn(netConn net.Conn) *Conn {
	cn := &Conn{
		netConn: netConn,
	}
	cn.SetUsedAt(time.Now())
	return cn
}

func (cn *Conn) GetConn() net.Conn {
	return cn.netConn
}

func (cn *Conn) UsedAt() time.Time {
	return cn.usedAt.Load().(time.Time)
}

func (cn *Conn) SetUsedAt(tm time.Time) {
	cn.usedAt.Store(tm)
}

func (cn *Conn) SetNetConn(netConn net.Conn) {
	cn.netConn = netConn
}

func (cn *Conn) IsStale(timeout time.Duration) bool {
	return timeout > 0 && time.Since(cn.UsedAt()) > timeout
}

func (cn *Conn) SetReadTimeout(timeout time.Duration) error {
	now := time.Now()
	cn.SetUsedAt(now)
	if timeout > 0 {
		return cn.netConn.SetReadDeadline(now.Add(timeout))
	}
	return cn.netConn.SetReadDeadline(noDeadline)
}

func (cn *Conn) SetWriteTimeout(timeout time.Duration) error {
	now := time.Now()
	cn.SetUsedAt(now)
	if timeout > 0 {
		return cn.netConn.SetWriteDeadline(now.Add(timeout))
	}
	return cn.netConn.SetWriteDeadline(noDeadline)
}

func (cn *Conn) Write(b []byte) (int, error) {
	return cn.netConn.Write(b)
}

func (cn *Conn) RemoteAddr() net.Addr {
	return cn.netConn.RemoteAddr()
}

func (cn *Conn) Close() error {
	return cn.netConn.Close()
}

func GetConn(connPool *ConnPool) (*Conn, bool, error) {
	cn, isNew, err := connPool.Get()
	if err != nil {
		return nil, false, err
	}
	return cn, isNew, nil
}

func (this *Conn) ReleaseConn(connPool *ConnPool, err error, callback func(error) bool) (bool, error) {
	if callback(err) {
		err = connPool.Remove(this)
		return false, err
	}

	err = connPool.Put(this)
	if err != nil {
		return false, err
	}
	return true, nil
}
