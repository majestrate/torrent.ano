package server

import (
	"io"
	"net"
	"net/textproto"
	"strings"
)

type serverConn struct {
	s        *Server
	c        *textproto.Conn
	group    string
	authUser string
	authed   bool
	mode     string
	open     bool
}

func newServerConn(s *Server, c net.Conn) *serverConn {
	return &serverConn{
		s:    s,
		c:    textproto.NewConn(c),
		mode: "READER",
		open: true,
	}
}

func (c *serverConn) run() {
	err := c.c.PrintfLine("201 Posting Not Allowed")
	if err != nil {
		c.Close()
	}
	for err == nil && c.open {
		var line string
		line, err = c.c.ReadLine()
		if err == nil {
			err = c.handleCMDLine(line, strings.ToUpper(line))
		}
	}
}

func (c *serverConn) sendCaps() error {
	c.c.PrintfLine("101 I can do the following")
	dw := c.c.DotWriter()
	for _, cap := range []string{"VERSION 2", "READER", "STREAMING", "IMPLEMENTATION torrent-index-nntpd", "POST", "IHAVE", "AUTHINFO", "STARTTLS"} {
		io.WriteString(dw, cap+"\n")
	}
	return dw.Close()
}

func (c *serverConn) Close() error {
	c.open = false
	return c.c.Close()
}

func (c *serverConn) handleCMDLine(line, upperLine string) (err error) {
	if upperLine == "CAPABILITIES" {
		return c.sendCaps()
	}
	if strings.HasPrefix(upperLine, "QUIT") {
		c.c.PrintfLine("205 kbai")
		return c.Close()
	}
	if strings.HasPrefix(upperLine, "AUTHINFO USER ") {
		c.authUser = strings.TrimSpace(line[14:])
		return c.c.PrintfLine("381 Password Required")
	}
	if strings.HasPrefix(upperLine, "AUTHINFO PASS ") {
		if len(c.authUser) == 0 {
			return c.c.PrintfLine("481 Auth Rejected")
		}
		passwd := strings.TrimSpace(line[14:])
		c.authed, _ = c.s.Auth.CheckLogin(c.authUser, passwd)
		msg := "481 Auth Rejected"
		if c.authed {
			msg = "281 Auth Accepted"
		}
		return c.c.PrintfLine(msg)
	}
	return c.c.PrintfLine("500 unknown command")
}
