package client

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestSocksRelayAuthenticatesUpstream(t *testing.T) {
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := upstream.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		buf := make([]byte, 3)
		if _, err = io.ReadFull(conn, buf); err != nil {
			done <- err
			return
		}
		if string(buf) != "\x05\x01\x02" {
			done <- &unexpected{string(buf)}
			return
		}
		_, _ = conn.Write([]byte{5, 2})
		if _, err = io.ReadFull(conn, buf[:2]); err != nil {
			done <- err
			return
		}
		body := make([]byte, int(buf[1])+1)
		if _, err = io.ReadFull(conn, body); err != nil {
			done <- err
			return
		}
		if string(body[:len(body)-1]) != "proxy" {
			done <- &unexpected{string(body)}
			return
		}
		pass := make([]byte, int(body[len(body)-1]))
		if _, err = io.ReadFull(conn, pass); err != nil {
			done <- err
			return
		}
		if string(pass) != "secret" {
			done <- &unexpected{string(pass)}
			return
		}
		_, _ = conn.Write([]byte{1, 0})
		done <- nil
	}()
	relay, err := StartSocksRelay("127.0.0.1:0", upstream.Addr().String(), "proxy:secret")
	if err != nil {
		t.Fatal(err)
	}
	defer relay.Close()
	conn, err := net.Dial("tcp", relay.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte{5, 1, 0})
	got := make([]byte, 2)
	if _, err = io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "\x05\x00" {
		t.Fatalf("client greeting = %v", got)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("upstream auth not reached")
	}
}

type unexpected struct{ got string }

func (e *unexpected) Error() string { return "unexpected bytes: " + e.got }
