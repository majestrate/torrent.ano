package server

import (
	"net"
	"usenet/auth"
	"usenet/store"
)

type Server struct {
	Store store.Store
	Auth  auth.Authenticator
}

func (s *Server) Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		conn := newServerConn(s, c)
		go conn.run()
	}
}
