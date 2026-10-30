package gos7

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
)

type statusTransport struct {
	response []byte
	err      error
}

func (s statusTransport) Send([]byte) ([]byte, error) { return s.response, s.err }

func TestPLCGetStatusShortResponse(t *testing.T) {
	for _, length := range []int{2, 28, 30, 31, 35, 44} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			client := NewClient2(&tcpPackager{}, statusTransport{response: make([]byte, length)})
			status, err := client.PLCGetStatus()
			if status != s7CpuStatusUnknown || err == nil || err.Error() != ErrorText(errIsoInvalidPDU) {
				t.Fatalf("PLCGetStatus = (%d, %v), want unknown and invalid PDU", status, err)
			}
		})
	}
}

func TestPLCGetStatusValues(t *testing.T) {
	for _, value := range []byte{0, 8, 4, 3} {
		response := make([]byte, 45)
		response[44] = value
		client := NewClient2(&tcpPackager{}, statusTransport{response: response})
		status, err := client.PLCGetStatus()
		want := int(value)
		if value == 3 {
			want = s7CpuStatusStop
		}
		if status != want || err != nil {
			t.Fatalf("status byte %d: PLCGetStatus = (%d, %v), want (%d, nil)", value, status, err, want)
		}
	}
}

func TestPLCGetStatusCPUError(t *testing.T) {
	response := make([]byte, 35)
	binary.BigEndian.PutUint16(response[27:], 0xd201)
	client := NewClient2(&tcpPackager{}, statusTransport{response: response})
	status, err := client.PLCGetStatus()
	want := ErrorText(CPUError(0xd201))
	if status != s7CpuStatusUnknown || err == nil || err.Error() != want {
		t.Fatalf("PLCGetStatus = (%d, %v), want unknown and %s", status, err, want)
	}
}

func TestPLCGetStatusTransportError(t *testing.T) {
	want := errors.New("connection lost")
	client := NewClient2(&tcpPackager{}, statusTransport{err: want})
	status, err := client.PLCGetStatus()
	if status != s7CpuStatusUnknown || err != want {
		t.Fatalf("PLCGetStatus = (%d, %v), want unknown and %v", status, err, want)
	}
}
