package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
)

//implement PLC hot start interface
func (mb *client) PLCHotStart() error {
	requestData := make([]byte, len(s7HotStartTelegram))
	copy(requestData, s7HotStartTelegram)
	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if err == nil {
		if length := len(response.Data); length > 18 { // 18 is the minimum expected
			if int(response.Data[19]) != pduStart {
				err = fmt.Errorf(ErrorText(errCliCannotStartPLC))
			} else {
				if int(response.Data[20]) == pduAlreadyStarted {
					err = fmt.Errorf(ErrorText(errCliAlreadyRun))
				} else {
					err = fmt.Errorf(ErrorText(errCliCannotStartPLC))
				}
			}
		} else {
			err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
		}
	}
	return err
}

//implement of PLC Colde Start interface
func (mb *client) PLCColdStart() error {
	requestData := make([]byte, len(s7ColdStartTelegram))
	copy(requestData, s7ColdStartTelegram)
	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if err == nil {
		if length := len(response.Data); length > 18 { // 18 is the minimum expected
			if int(response.Data[19]) != pduStart {
				err = fmt.Errorf(ErrorText(errCliCannotStartPLC))
			} else {
				if int(response.Data[20]) == pduAlreadyStarted {
					err = fmt.Errorf(ErrorText(errCliAlreadyRun))
				} else {
					err = fmt.Errorf(ErrorText(errCliCannotStartPLC))
				}
			}
		} else {
			err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
		}
	}
	return err
}
func (mb *client) PLCStop() error {
	requestData := make([]byte, len(s7StopTelegram))
	copy(requestData, s7StopTelegram)

	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if err == nil {
		if length := len(response.Data); length > 18 { // 18 is the minimum expected
			if int(response.Data[19]) != pduStop {
				err = fmt.Errorf(ErrorText(errCliCannotStopPLC))
			} else {
				if int(response.Data[20]) == pduAlreadyStarted {
					err = fmt.Errorf(ErrorText(errCliAlreadyStop))
				} else {
					err = fmt.Errorf(ErrorText(errCliCannotStopPLC))
				}
			}
		} else {
			err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
		}
	}
	return err
}

//
func (mb *client) PLCGetStatus() (status int, err error) {
	//initialize
	requestData := make([]byte, len(s7GetStatusTelegram))
	copy(requestData, s7GetStatusTelegram)

	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if err == nil {
		if length := len(response.Data); length > 30 { // 30 is the minimum expected
			if result := binary.BigEndian.Uint16(response.Data[27:]); result == 0 {
				if int(response.Data[44]) == 0 || int(response.Data[44]) == 8 || int(response.Data[44]) == 4 {
					status = int(response.Data[44])
				} else {
					// Since RUN status is always 8 for all CPUs and CPs, STOP status
					// sometime can be coded as 3 (especially for old cpu...)
					status = s7CpuStatusStop
				}

			} else {
				err = fmt.Errorf(ErrorText(CPUError(uint(result))))
			}
		} else {
			err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
		}
	}
	return
}
