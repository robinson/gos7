package gos7

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

// putGetDisabledReply is the reply of an S7-1200 with PUT/GET access disabled
// to a read request: an Ack with error class 0x81, code 0x04 and no data.
var putGetDisabledReply = []byte{
	0x03, 0x00, 0x00, 0x13, // TPKT
	0x02, 0xF0, 0x80, // COTP
	0x32, 0x02, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x81, 0x04,
}

func TestResponseError(t *testing.T) {
	ackData := func(class, code byte) []byte {
		data := append([]byte(nil), putGetDisabledReply...)
		data[8], data[17], data[18] = s7PDUTypeAckData, class, code
		return append(data, 0x04, 0x01) // read var parameter
	}
	withType := func(rosctr byte) []byte {
		data := ackData(0x81, 0x04)
		data[8] = rosctr
		return data
	}

	tests := []struct {
		name string
		data []byte
		want *S7Error
	}{
		{"Ack with error", putGetDisabledReply, &S7Error{High: 0x81, Low: 0x04}},
		{"Ack_Data with error", ackData(0x85, 0x00), &S7Error{High: 0x85, Low: 0x00}},
		{"Ack_Data without error", ackData(0, 0), nil},
		{"Job has no error fields", withType(0x01), nil},
		{"Userdata has no error fields", withType(0x07), nil},
		{"not an S7 PDU", make([]byte, 30), nil},
		{"shorter than the Ack header", putGetDisabledReply[:18], nil},
		{"empty", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := responseError(&ProtocolDataUnit{Data: tt.data})
			if tt.want == nil {
				if err != nil {
					t.Fatalf("responseError = %v, want nil", err)
				}
				return
			}
			var s7Err *S7Error
			if !errors.As(err, &s7Err) || *s7Err != *tt.want {
				t.Fatalf("responseError = %#v, want %#v", err, tt.want)
			}
		})
	}
}

// TestAGReadDBPLCErrorResponse reads through a real TCPClientHandler from a
// PLC that rejects the request. Before, readArea reported the 19 byte reply as
// "ISO : Invalid Buffer passed to Send/Receive'19'".
func TestAGReadDBPLCErrorResponse(t *testing.T) {
	clientEnd, plcEnd := net.Pipe()
	defer clientEnd.Close()
	defer plcEnd.Close()

	go func() {
		header := make([]byte, 4)
		if _, err := io.ReadFull(plcEnd, header); err != nil {
			return
		}
		request := make([]byte, binary.BigEndian.Uint16(header[2:])-4)
		if _, err := io.ReadFull(plcEnd, request); err != nil {
			return
		}
		reply := append([]byte(nil), putGetDisabledReply...)
		// Echo the PDU reference. request starts after the 4 TPKT bytes.
		copy(reply[s7PDURefOffset:s7PDURefOffset+2], request[s7PDURefOffset-4:])
		_, _ = plcEnd.Write(reply)
	}()

	handler := &TCPClientHandler{}
	handler.conn = clientEnd
	handler.Timeout = time.Second
	handler.PDULength = 240
	client := NewClient(handler)

	err := client.AGReadDB(250, 4, 5, make([]byte, 5))

	var s7Err *S7Error
	if !errors.As(err, &s7Err) || s7Err.High != 0x81 || s7Err.Low != 0x04 {
		t.Fatalf("AGReadDB = %v, want S7Error 0x81 0x04", err)
	}
}
