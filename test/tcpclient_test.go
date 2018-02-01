package test

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"log"
	"os"
	"testing"
	"time"

	"../../gos7"
)

const (
	tcpDevice = "192.168.72.129"
	rack      = 0
	slot      = 2
)

func TestTCPClient(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 200 * time.Second
	handler.IdleTimeout = 200 * time.Second
	handler.Logger = log.New(os.Stdout, "tcp: ", log.LstdFlags)
	handler.Connect()
	defer handler.Close()

	client := gos7.NewClient(handler)

	ClientTestAll(t, client)
}
