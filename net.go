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

func (this *WaitListener) Accept() (conn net.Conn, err error) {

	//Call Add before the event to be waited for (the connection)
	this.WaitGroup.Add(1)
	//Call the underlying listener's Accept() method.
	c, err := this.Listener.Accept()
	if err != nil {
		this.WaitGroup.Done()
		return
	}

	//Wrap the connection in a WaitConn.
	conn = &WaitConn{
		Conn: c,
		//Use the WaitListener's WaitGroup
		WaitGroup: this.WaitGroup,
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

func (this *WaitConn) Close() error {
	//Is possible to call Close() more than once?
	defer this.once.Do(this.WaitGroup.Done)
	//Close the underlying connection.
	return this.Conn.Close()
}
