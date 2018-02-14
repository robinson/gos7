package test

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"runtime"
	"strings"
	"testing"

	"../../gos7"
)

//ClientTestAll client test all
func ClientTestAll(t *testing.T, client gos7.Client) {
	//write value to 100
	ClientTestWriteIntDB(t, client, 100)
	//read and assert with 100
	ClientTestReadIntDB(t, client)
	//return 0
	ClientTestWriteIntDB(t, client, 0)
	//test directory
	ClientTestDirectory(t, client)
	//Get CPU info
	ClientTestGetCPUInfo(t, client)
	//Get AG Block Info
	ClientTestGetAGBlockInfo(t, client)
}

//ClientTestWriteIntDB client test write int
func ClientTestWriteIntDB(t *testing.T, client gos7.Client, value int16) {
	address := 2710
	start := 8
	size := 2
	buffer := make([]byte, 255)

	//binary.BigEndian.PutUint16(buffer[0:], uint16(value))
	var helper gos7.Helper
	helper.SetValueAt(buffer, 0, value)
	err := client.AGWriteDB(address, start, size, buffer)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, err, nil) // send success then the response in position 6 will be 128
}

//ClientTestReadIntDB client test read int
func ClientTestReadIntDB(t *testing.T, client gos7.Client) {
	address := 2710
	start := 8
	size := 2
	buf := make([]byte, 255)
	err := client.AGReadDB(address, start, size, buf)
	if err != nil {
		t.Fatal(err)
	}
	// result := binary.BigEndian.Uint16(results)
	var s7 gos7.Helper
	var result uint16
	s7.GetValueAt(buf, 0, &result)
	AssertEquals(t, 100, int(result))
}

//ClientTestDirectory test directory functions, list all blocks
func ClientTestDirectory(t *testing.T, client gos7.Client) {
	//var bl gos7.S7BlocksList
	bl, err := client.PGListBlocks()
	if err != nil {
		t.Fatal(err)
	}
	//for example
	AssertEquals(t, len(bl.OBList), 10)
	AssertEquals(t, len(bl.DBList), 113)
	AssertEquals(t, len(bl.FBList), 81)
}

//ClientTestGetCPUInfo get the CPU info
func ClientTestGetCPUInfo(t *testing.T, client gos7.Client) {
	info, err := client.GetCPUInfo()
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, strings.Index(info.SerialNumber, "0118701484"), 0) //return serial should be "0118701484        ", some space
}

//ClientTestGetAGBlockInfo get AG block info
func ClientTestGetAGBlockInfo(t *testing.T, client gos7.Client) {
	info, err := client.GetAgBlockInfo(65, 2710)
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, info.CodeDate, "22.01.2018")
}

//AssertEquals helper
func AssertEquals(t *testing.T, expected, actual interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	} else {
		// Get file name only
		idx := strings.LastIndex(file, "/")
		if idx >= 0 {
			file = file[idx+1:]
		}
	}

	if expected != actual {
		t.Logf("%s:%d: Expected: %+v (%T), actual: %+v (%T)", file, line,
			expected, expected, actual, actual)
		t.FailNow()
	}
}
