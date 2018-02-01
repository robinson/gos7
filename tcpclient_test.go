package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestTCPTransporter(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, err = io.Copy(conn, conn)
		if err != nil {
			t.Error(err)
			return
		}
	}()
	client := &tcpTransporter{
		Address:     ln.Addr().String(),
		Timeout:     200 * time.Second,
		IdleTimeout: 100 * time.Millisecond,
	}
	req := []byte{0, 1, 0, 17, 0, 2, 1, 2, 0, 1, 0, 17, 0, 2, 1, 2, 2} //lengh 17, > MinPduSize

	client.tcpConnect() //assume tcp connect to test locally
	rsp, err := client.Send(req)
	if err != nil {
		t.Fatal(err)
	}
	//lth: just compare 7 first byte
	if !bytes.Equal(req, rsp) {
		t.Fatalf("unexpected response: %x", rsp)
	}
	time.Sleep(150 * time.Millisecond)
	if client.conn != nil {
		t.Fatalf("connection is not closed: %+v", client.conn)
	}
}
