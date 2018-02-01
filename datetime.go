package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
	"time"
)

//implement GetPLCDateTime
func (mb *client) PGClockWrite() (datetime time.Time, err error) {
	requestData := make([]byte, len(s7GetDatetimeTelegram))
	copy(requestData, s7GetDatetimeTelegram)
	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if length := len(response.Data); length > 30 {
		if (binary.BigEndian.Uint16(response.Data[27:]) == 0) && (response.Data[29] == 0xFF) {
			var s7 Helper
			datetime = s7.GetDateTimeAt(response.Data, 35)
		} else {
			err = fmt.Errorf(ErrorText(errCliInvalidPlcAnswer))
		}

	} else {
		err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
	}
	return
}

//implement SetPLCDateTime
func (mb *client) PGClockRead(datetime time.Time) (err error) {
	requestData := make([]byte, len(s7SetDatetimeTelegram))
	copy(requestData, s7SetDatetimeTelegram)
	var s7 Helper
	s7.SetDateTimeAt(requestData, 32, datetime)

	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if length := len(response.Data); length > 30 {
		if binary.BigEndian.Uint16(response.Data[27:]) != 0 {
			err = fmt.Errorf(ErrorText(errCliInvalidPlcAnswer))
		}
	} else {
		err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
	}
	return
}
