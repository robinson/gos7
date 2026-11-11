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
	if err != nil {
		return
	}
	if length := len(response.Data); length > 30 {
		if (binary.BigEndian.Uint16(response.Data[27:]) == 0) && (response.Data[29] == 0xFF) {
			if length < 43 { // A successful response must carry the 10 byte time field at 33.
				err = fmt.Errorf("%s", ErrorText(errIsoInvalidPDU))
				return
			}
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
	// The 10 byte time field at 29 holds a reserved byte, the century and the
	// 8 byte DATE_AND_TIME. Writing at 32 shifted every field by one and ran
	// past the end of the 39 byte telegram.
	requestData[30] = encodeBcd(datetime.Year() / 100)
	var s7 Helper
	s7.SetDateTimeAt(requestData, 31, datetime)

	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	if length := len(response.Data); length > 30 {
		if binary.BigEndian.Uint16(response.Data[27:]) != 0 {
			err = fmt.Errorf(ErrorText(errCliInvalidPlcAnswer))
		}
	} else {
		err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
	}
	return
}
