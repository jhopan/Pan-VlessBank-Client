package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type SocksRelay struct {
	listener net.Listener
	upstream string
	user     string
	pass     string
	wg       sync.WaitGroup
}

func StartSocksRelay(listen, upstream, auth string) (*SocksRelay, error) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok || user == "" || pass == "" || len(user) > 255 || len(pass) > 255 {
		return nil, errors.New("auth proxy harus USER:PASSWORD")
	}
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}
	r := &SocksRelay{listener: listener, upstream: upstream, user: user, pass: pass}
	r.wg.Add(1)
	go r.serve()
	return r, nil
}

func (r *SocksRelay) Addr() net.Addr { return r.listener.Addr() }

func (r *SocksRelay) Close() error {
	err := r.listener.Close()
	r.wg.Wait()
	return err
}

func (r *SocksRelay) serve() {
	defer r.wg.Done()
	for {
		client, err := r.listener.Accept()
		if err != nil {
			return
		}
		r.wg.Add(1)
		go func() { defer r.wg.Done(); r.handle(client) }()
	}
}

func (r *SocksRelay) handle(client net.Conn) {
	defer client.Close()
	if err := clientGreeting(client); err != nil {
		return
	}
	upstream, err := net.Dial("tcp", r.upstream)
	if err != nil {
		return
	}
	defer upstream.Close()
	if err := upstreamAuth(upstream, r.user, r.pass); err != nil {
		return
	}
	copyDone := make(chan struct{})
	go func() { _, _ = io.Copy(upstream, client); _ = upstream.Close(); close(copyDone) }()
	_, _ = io.Copy(client, upstream)
	<-copyDone
}

func clientGreeting(conn net.Conn) error {
	var head [2]byte
	if _, err := io.ReadFull(conn, head[:]); err != nil || head[0] != 5 {
		return errors.New("SOCKS client greeting tidak valid")
	}
	methods := make([]byte, head[1])
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}
	for _, method := range methods {
		if method == 0 {
			_, err := conn.Write([]byte{5, 0})
			return err
		}
	}
	_, _ = conn.Write([]byte{5, 255})
	return errors.New("client tidak mendukung SOCKS tanpa auth")
}

func upstreamAuth(conn net.Conn, user, pass string) error {
	if _, err := conn.Write([]byte{5, 1, 2}); err != nil {
		return err
	}
	var method [2]byte
	if _, err := io.ReadFull(conn, method[:]); err != nil || method != [2]byte{5, 2} {
		return errors.New("bridge SOCKS tidak menerima auth")
	}
	request := make([]byte, 0, len(user)+len(pass)+3)
	request = append(request, 1, byte(len(user)))
	request = append(request, user...)
	request = append(request, byte(len(pass)))
	request = append(request, pass...)
	if _, err := conn.Write(request); err != nil {
		return err
	}
	var response [2]byte
	if _, err := io.ReadFull(conn, response[:]); err != nil || response != [2]byte{1, 0} {
		return fmt.Errorf("auth proxy ditolak")
	}
	return nil
}
