package acceptor

import (
	"syscall"

	ep "github.com/halivor/goutility/eventpool"
	log "github.com/halivor/goutility/logger"
	sc "github.com/halivor/sunshine/conf"
	c "github.com/halivor/sunshine/connection"
	p "github.com/halivor/sunshine/peer"
)

type Acceptor struct {
	ev   ep.EP_EVENT
	addr string

	c.Conn
	ep.EventPool // event: add, del, mod
	p.Manager

	log.Logger
}

func NewTcpAcceptor(addr string, epr ep.EventPool, mgr p.Manager) *Acceptor {
	C, e := c.NewTcpConn()
	if e != nil {
		panic(e)
	}
	saddr, e := c.ParseAddr4("tcp", addr)
	if e != nil {
		panic(e)
	}
	if e := c.Reuse(C.Fd(), true); e != nil {
		log.Warn("reuse addr port failed,", e)
	}
	if e = syscall.Bind(C.Fd(), saddr); e != nil {
		panic(e)
	}
	if e = syscall.Listen(C.Fd(), 1024); e != nil {
		panic(e)
	}

	a := &Acceptor{
		ev:        ep.EV_READ,
		addr:      addr,
		Conn:      C,
		EventPool: epr,
		Manager:   mgr,
		Logger:    log.NewLog("sunsine.log", "[ac]", log.LstdFlags, sc.LogLvlAccept),
	}
	a.AddEvent(a)

	return a
}

func (a *Acceptor) CallBack(ev ep.EP_EVENT) {
	if ev&ep.EV_ERROR != 0 {
		a.Warn("epoll error", ev)
		a.DelEvent(a)
		return
	}
	switch fd, _, e := syscall.Accept4(a.Fd(), syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC); e {
	case syscall.EAGAIN, syscall.EINTR:
	case syscall.EMFILE, syscall.ENFILE, syscall.ENOBUFS, syscall.ENOMEM:
		// EMFILE => The per-process limit of open file descriptors has been reached.
		// ENFILE => The system limit on the total number of open files has been reached.
		// ENOBUFS, ENOMEM => Not enough free memory. This often means that the memory allocation is limited by the socket buffer limits, not by the system memory.
	case nil:
		a.Debug("new peer", fd)
		a.AddEvent(p.New(c.NewConn(fd), a.EventPool, a.Manager))
	default:
		a.DelEvent(a)
		// TODO: 释放并重启acceptor
	}
}
func (a *Acceptor) Event() ep.EP_EVENT {
	return a.ev
}
func (a *Acceptor) Release() {
	a.Conn.Close()
}
