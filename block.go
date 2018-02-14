package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

import (
	"encoding/binary"
	"fmt"
	"time"
)

// S7BlockInfo Managed Block Info
type S7BlockInfo struct {
	BlkType   int
	BlkNumber int
	BlkLang   int
	BlkFlags  int
	MC7Size   int // The real size in bytes
	LoadSize  int
	LocalData int
	SBBLength int
	CheckSum  int
	Version   int
	// Chars info
	CodeDate string
	IntfDate string
	Author   string
	Family   string
	Header   string
}

func (mb *client) DBFill(dbnumber int, fillChar int) (err error) {
	// bi := S7BlockInfo{}
	bi, err := mb.GetAgBlockInfo(blockDB, dbnumber)
	if err == nil {
		buffer := make([]byte, bi.MC7Size)
		for c := 0; c < bi.MC7Size; c++ {
			buffer[c] = byte(fillChar)
		}
		err = mb.AGWriteDB(dbnumber, 0, bi.MC7Size, buffer)
	}
	return
}

func (mb *client) DBGet(dbnumber int, usrdata []byte, size int) (err error) {
	// bi := S7BlockInfo{}
	bi, err := mb.GetAgBlockInfo(blockDB, dbnumber)
	if err == nil {
		if dbSize := bi.MC7Size; dbSize <= len(usrdata) {
			size = dbSize
			err = mb.AGReadDB(dbnumber, 0, dbSize, usrdata)
			if err == nil {
				size = dbSize
			}
		} else {
			err = fmt.Errorf(ErrorText(errCliBufferTooSmall))
		}
	}
	return
}

//internal class returns info about a given block in PLC memory.
//This function is very useful if you need to read or write data in a DB
//which you do not know the size in advance ( MC7Size).
func (mb *client) GetAgBlockInfo(blocktype int, blocknum int) (info S7BlockInfo, err error) {
	//init buffer
	requestData := make([]byte, len(s7BlockInfoTelegram))
	copy(requestData, s7BlockInfoTelegram)
	requestData[30] = byte(blocktype)
	// Block Number
	requestData[31] = byte((blocknum / 10000) + 0x30)
	blocknum = blocknum % 10000
	requestData[32] = byte((blocknum / 1000) + 0x30)
	blocknum = blocknum % 1000
	requestData[33] = byte((blocknum / 100) + 0x30)
	blocknum = blocknum % 100
	requestData[34] = byte((blocknum / 10) + 0x30)
	blocknum = blocknum % 10
	requestData[35] = byte((blocknum / 1) + 0x30)
	request := NewProtocolDataUnit(requestData)
	//send
	response, err := mb.send(&request)
	if err == nil {
		if length := len(response.Data); length > 32 {
			if result := binary.BigEndian.Uint16(response.Data[27:]); result == 0 {
				info.BlkFlags = int(response.Data[42])
				info.BlkLang = int(response.Data[43])
				info.BlkType = int(response.Data[44])
				info.BlkNumber = int(binary.BigEndian.Uint16(response.Data[45:]))
				info.LoadSize = int(binary.BigEndian.Uint32(response.Data[47:]))
				info.CodeDate = siemensTimestamp(int64(binary.BigEndian.Uint16(response.Data[59:])))
				info.IntfDate = siemensTimestamp(int64(binary.BigEndian.Uint16(response.Data[65:])))
				info.SBBLength = int(binary.BigEndian.Uint16(response.Data[67:]))
				info.LocalData = int(binary.BigEndian.Uint16(response.Data[71:]))
				info.MC7Size = int(binary.BigEndian.Uint16(response.Data[73:]))
				info.Author = string(response.Data[75 : 75+8])
				info.Family = string(response.Data[83 : 83+8])
				info.Header = string(response.Data[91 : 91+8])
				info.Version = int(response.Data[99])
				info.CheckSum = int(binary.BigEndian.Uint16(response.Data[101:]))
			} else {
				err = fmt.Errorf(ErrorText(CPUError(uint(result))))
			}

		} else {
			err = fmt.Errorf(ErrorText(errIsoInvalidPDU))
		}
	}
	return
}

//siemensTimestamp helper get Siemens timestamp
func siemensTimestamp(EncodedDate int64) string {
	return time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Second * time.Duration((EncodedDate * 86400))).Format("02.01.2006")
}
