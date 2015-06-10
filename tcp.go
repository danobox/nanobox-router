package router

import (
	"errors"
	"io"
	"net"
)

func (r *Router) AddForward(remote string) (int, error) {
	laddr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 0,
	}

	listener, err := net.ListenTCP("tcp4", &laddr)
	if err != nil {
		return 0, err
	}

	r.Forwards[remote] = listener

	go r.tcpListen(listener, remote)

	tcpAddr, _ := net.ResolveTCPAddr("tcp", listener.Addr().String())

	return tcpAddr.Port, nil
}

func (r *Router) GetForward(remote string) *net.TCPListener {
	return r.Forwards[remote]
}

func (r *Router) GetLocalPort(remote string) int {
	listener := r.Forwards[remote]
	if listener == nil {
		return 0
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", listener.Addr().String())

	return tcpAddr.Port
}

func (r *Router) RemoveForward(remote string) error {
	listener := r.Forwards[remote]
	if listener == nil {
		return errors.New("I could not find the forward you seek")
	}
	listener.Close()
	delete(r.Forwards, remote)
	return nil
}

func (r *Router) tcpListen(listener *net.TCPListener, remoteAddr string) {
	for {
		conn, err := listener.Accept()
		if conn == nil {
			r.handleError(err)
			return
		}
		go r.tcpForward(conn, remoteAddr)
	}
}

func (r *Router) tcpForward(local net.Conn, remoteAddr string) {
	remote, err := net.Dial("tcp", remoteAddr)
	if remote == nil {
		r.handleError(err)
		local.Close()
		return
	}
	go func() {
		defer local.Close()
		// remote.SetReadTimeout(120*1E9)
		io.Copy(local, remote)
	}()
	go func() {
		defer remote.Close()
		//local.SetReadTimeout(120*1E9)
		io.Copy(remote, local)
	}()

}