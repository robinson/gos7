package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
	"strings"
)

//SZLHeader See §33.1 of "System Software for S7-300/400 System and Standard Functions" and see SFC51 description too
type SZLHeader struct {
	LengthHeader       uint16
	NumberOfDataRecord uint16
}

//S7SZL constains header and data
type S7SZL struct {
	Header SZLHeader
	Data   []byte
}

// S7SZLList of available SZL IDs : same as SZL but List items are big-endian adjusted
type S7SZLList struct {
	Header SZLHeader
	Data   []uint16
}

// S7Protection See §33.19 of "System Software for S7-300/400 System and Standard Functions"
type S7Protection struct {
	schSchal uint // sch_schal: Protection level set with the mode selector (1, 2, 3)
	schPar   uint // sch_par: Protection level set in parameters (0, 1, 2, 3; 0: no password,protection level invalid)
	schRel   uint // sch_rel: Valid protection level of the CPU
	bartSch  uint // bart_sch: Mode selector setting (1:RUN, 2:RUN-P, 3:STOP, 4:MRES,0:undefined or cannot be determined)
	anlSch   uint // anl_sch:Startup switch setting (1:CRST, 2:WRST, 0:undefined, does not exist of cannot be determined)
}

//S7OrderCode Order Code + Version
type S7OrderCode struct {
	Code string // such as "6ES7 151-8AB01-0AB0"
	V1   byte   // Version 1st digit
	V2   byte   // Version 2nd digit
	V3   byte   // Version 3th digit
}

//S7CpuInfo CPU Info
type S7CpuInfo struct {
	ModuleTypeName string
	SerialNumber   string
	ASName         string
	Copyright      string
	ModuleName     string
}

//S7CpInfo cp info
type S7CpInfo struct {
	MaxPduLength   int
	MaxConnections int
	MaxMpiRate     int
	MaxBusRate     int
}

//implement GetCPUInfo
func (mb *client) GetCPUInfo() (info S7CpuInfo, err error) {

	szl, _, err := mb.readSzl(0x001C, 0x000)
	if err == nil {
		moduleTypeName := string(szl.Data[172 : 172+32])
		serialNumber := string(szl.Data[138 : 138+24])
		asName := string(szl.Data[2 : 2+24])
		copyRight := string(szl.Data[104 : 104+26])
		moduleName := string(szl.Data[36 : 36+24])

		info.ModuleTypeName = strings.TrimSpace(moduleTypeName)
		info.SerialNumber = strings.TrimSpace(serialNumber)
		info.ASName = strings.TrimSpace(asName)
		info.Copyright = strings.TrimSpace(copyRight)
		info.ModuleName = strings.TrimSpace(moduleName)
	}
	return
}

//implement of GetCPInfo
func (mb *client) GetCPInfo() (info S7CpInfo, err error) {
	szl, _, err := mb.readSzl(0x0131, 0x000)
	if err == nil {
		info.MaxPduLength = int(binary.BigEndian.Uint16(szl.Data[2:]))
		info.MaxConnections = int(binary.BigEndian.Uint16(szl.Data[4:]))
		info.MaxMpiRate = int(binary.BigEndian.Uint16(szl.Data[6:]))
		info.MaxBusRate = int(binary.BigEndian.Uint16(szl.Data[10:]))
	}
	return
}

//implement of GetOrderCode
func (mb *client) GetOrderCode() (info S7OrderCode, err error) {
	szl, size, err := mb.readSzl(0x0131, 0x000)
	if err == nil {
		info.Code = string(szl.Data[2 : 2+20])
		info.V1 = szl.Data[size-3]
		info.V2 = szl.Data[size-2]
		info.V3 = szl.Data[size-1]
	}
	return
}

//internal function readSZL
func (mb *client) readSzl(id int, index int) (szl S7SZL, size int, err error) {
	var dataSZL int
	offset := 0
	var done bool
	first := true
	var seqIn byte = 0x00
	var seqOut uint16 = 0x0000
	// szl = S7SZL{	}
	// szl.Header.LengthHeader = 0
	s7SZLFirst := make([]byte, len(s7SZLFirstTelegram))
	copy(s7SZLFirst, s7SZLFirstTelegram)
	s7SZLNext := make([]byte, len(s7SZLNextTelegram))
	copy(s7SZLNext, s7SZLNextTelegram)
	for !done && err == nil {
		res := &ProtocolDataUnit{}
		if first == true {
			binary.BigEndian.PutUint16(s7SZLFirst[11:], seqOut+1)
			binary.BigEndian.PutUint16(s7SZLFirst[29:], uint16(id))
			binary.BigEndian.PutUint16(s7SZLFirst[31:], uint16(index))
			request := NewProtocolDataUnit(s7SZLFirst)
			//send
			res, err = mb.send(&request)
		} else {
			binary.BigEndian.PutUint16(s7SZLNext[11:], seqOut+1)
			s7SZLNext[24] = byte(seqIn)
			request := NewProtocolDataUnit(s7SZLNext)
			//send
			res, err = mb.send(&request)
		}
		if err != nil {
			return
		}
		if length := len(res.Data); length <= 32 {
			err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
			return
		}
		if binary.BigEndian.Uint16(res.Data[27:]) != 0 && res.Data[29] != byte(0xFF) {
			err = fmt.Errorf(ErrorText(errCliInvalidPlcAnswer))
			return
		}
		if first {
			// Gets Amount of this slice
			dataSZL = int(binary.BigEndian.Uint16(res.Data[31:])) - 8 // Skips extra params (ID, Index ...)
			done = res.Data[26] == 0x00
			seqIn = byte(res.Data[24]) // Slice sequence
			//header
			header := SZLHeader{}
			header.LengthHeader = binary.BigEndian.Uint16(res.Data[37:])
			header.NumberOfDataRecord = binary.BigEndian.Uint16(res.Data[39:])
			//data
			data := make([]byte, offset+dataSZL)
			copy(data[offset:offset+dataSZL], res.Data[41:41+dataSZL])
			//s7szl
			szl.Header = header
			szl.Data = data

			offset += dataSZL
			szl.Header.LengthHeader += szl.Header.LengthHeader
		} else {
			dataSZL = int(binary.BigEndian.Uint16(res.Data[31:]))
			done = res.Data[26] == 0x00
			seqIn = byte(res.Data[24]) // Slice sequence
			data := make([]byte, offset+dataSZL)
			szl.Data = data

			copy(szl.Data[offset:offset+dataSZL], res.Data[37:37+dataSZL])
			offset += dataSZL
			szl.Header.LengthHeader += szl.Header.LengthHeader
		}
		first = false
	}
	return szl, size, err
}
