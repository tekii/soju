package soju

import (
	"net"
	"sync"
)

//type WaitListener wraps a net.Listener and has a sync.WaitGroup to track
//all returned connections.
type WaitListener struct {
	net.Listener
	WaitGroup *sync.WaitGroup
}

func (wl *WaitListener) Accept() (conn net.Conn, err error) {

	//Call Add before the event to be waited for (the connection)
	wl.WaitGroup.Add(1)
	//Call the underlying listener's Accept() method.
	c, err := wl.Listener.Accept()
	if err != nil {
		wl.WaitGroup.Done()
		return
	}

	//Wrap the connection in a WaitConn.
	conn = &WaitConn{
		Conn: c,
		//Use the WaitListener's WaitGroup
		WaitGroup: wl.WaitGroup,
	}

	return

}

//type WaitConn wraps a net.Conn and decrements the received WaitGroup
//when the connection is closed.
type WaitConn struct {
	net.Conn
	WaitGroup *sync.WaitGroup
	once      sync.Once
}

func (wc *WaitConn) Close() error {
	//Is possible to call Close() more than once?
	defer wc.once.Do(wc.WaitGroup.Done)
	//Close the underlying connection.
	return wc.Conn.Close()
}
