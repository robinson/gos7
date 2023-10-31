package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
)

//S7DataItem which expose as S7DataItem to use in Multiple read/write
type S7DataItem struct {
	Area     int
	WordLen  int
	DBNumber int
	Start    int
	Bit      int
	Amount   int
	Data     []byte
	Error    string
}

//implement WriteMulti
func (mb *client) AGWriteMulti(dataItems []S7DataItem, itemsCount int) (err error) {
	// Checks items
	if itemsCount > 20 { //max variable is 20
		err = fmt.Errorf(ErrorText(errCliTooManyItems))
		return
	}
	//fills header
	s7Multi := make([]byte, len(s7MultiWriteHeaderTelegram))
	copy(s7Multi, s7MultiWriteHeaderTelegram)

	parLength := itemsCount*len(s7MultiWriteItemTelegram) + 2
	binary.BigEndian.PutUint16(s7Multi[13:], uint16(parLength))
	s7Multi[18] = byte(itemsCount)
	// Fills Params
	offset := len(s7MultiWriteHeaderTelegram)
	for i := 0; i < itemsCount; i++ {
		s7ParamItem := make([]byte, len(s7MultiWriteItemTelegram))
		copy(s7ParamItem, s7MultiWriteItemTelegram)
		s7ParamItem[3] = byte(dataItems[i].WordLen)                                //word length
		s7ParamItem[8] = byte(dataItems[i].Area)                                   //area
		binary.BigEndian.PutUint16(s7ParamItem[4:], uint16(dataItems[i].Amount))   //amount
		binary.BigEndian.PutUint16(s7ParamItem[6:], uint16(dataItems[i].DBNumber)) //DBNo

		// Adjusts the offset
		var addr int
		if dataItems[i].WordLen == s7wlbit || dataItems[i].WordLen == s7wlcounter || dataItems[i].WordLen == s7wltimer {
			addr = dataItems[i].Start
		} else {
			addr = dataItems[i].Start * 8
		}

		// Build the offset
		s7ParamItem[11] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7ParamItem[10] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7ParamItem[9] = byte(addr & 0x0FF)
		// copy(s7Multi[offset:offset+len(s7ParamItem)], s7ParamItem[0:])
		s7Multi = append(s7Multi[:offset], append(s7ParamItem, s7Multi[offset:]...)...)
		offset += len(s7ParamItem)
	}
	dataLength := 0
	for i := 0; i < itemsCount; i++ {
		s7ItemWrite := make([]byte, 1024)
		s7ItemWrite[0] = 0
		itemDataSize := 0
		switch dataItems[i].WordLen {
		case s7wlbit:
			s7ItemWrite[1] = tsResBit
			itemDataSize = dataItems[i].Amount
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize))
			break
		case s7wlcounter:
		case s7wltimer:
			s7ItemWrite[1] = tsResOctet
			itemDataSize = dataItems[i].Amount * 2
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize))
			break
		case s7wlreal:
			s7ItemWrite[1] = tsResReal // real
			itemDataSize = dataItems[i].Amount * dataSizeByte(dataItems[i].WordLen)
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize))
			break
		default:
			s7ItemWrite[1] = tsResByte // byte/word/dword etc.
			itemDataSize = dataItems[i].Amount * dataSizeByte(dataItems[i].WordLen)
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize*8))
			break

		}
		copy(s7ItemWrite[4:4+itemDataSize], dataItems[i].Data)
		if itemDataSize%2 != 0 {
			s7ItemWrite[itemDataSize+4] = 0
			itemDataSize++
		}
		// copy(s7Multi[offset:offset+itemDataSize+4], s7ItemWrite[0:itemDataSize+4])
		s7Multi = append(s7Multi, s7ItemWrite[0:itemDataSize+4]...)
		offset = offset + itemDataSize + 4
		dataLength = dataLength + itemDataSize + 4
	}
	tt, _ := interface{}(mb.transporter).(*TCPClientHandler)
	//Checks the size
	if offset > tt.PDULength {
		err = fmt.Errorf(ErrorText(errCliSizeOverPDU))
		return
	}
	binary.BigEndian.PutUint16(s7Multi[2:], uint16(offset))      // Whole size
	binary.BigEndian.PutUint16(s7Multi[15:], uint16(dataLength)) // Whole size
	request := NewProtocolDataUnit(s7Multi)
	//debug
	fmt.Printf("%d", s7Multi)
	//send
	response, err := mb.send(&request)
	if err == nil {
		// Check Global Operation Result
		cpuErr := CPUError(uint(binary.BigEndian.Uint16(response.Data[17:])))
		if cpuErr != 0 {
			err = fmt.Errorf(ErrorText(cpuErr))
			return
		}
		if itemsWritten := int(response.Data[20]); itemsWritten != itemsCount || itemsWritten > 20 { //max var = 20
			err = fmt.Errorf(ErrorText(errCliInvalidPlcAnswer))
			return
		}
		for i := 0; i < itemsCount; i++ {
			if response.Data[i+21] == 0xFF {

				dataItems[i].Error = ""
			} else {
				dataItems[i].Error = ErrorText(CPUError(uint(response.Data[i+21])))
			}
		}
	}
	return
}

//implement ReadMulti
func (mb *client) AGReadMulti(dataItems []S7DataItem, itemsCount int) (err error) {
	// Checks items
	if itemsCount > 20 { //max variable is 20
		err = fmt.Errorf(ErrorText(errCliTooManyItems))
		return
	}
	s7Item := make([]byte, 12)
	s7Multi := make([]byte, len(s7MultiReadHeaderTelegram))
	copy(s7Multi, s7MultiReadHeaderTelegram)
	// Fills Header
	binary.BigEndian.PutUint16(s7Multi[13:], uint16(itemsCount*len(s7Item)+2))
	s7Multi[18] = byte(itemsCount)
	// Fills the Items
	offset := 19
	for i := 0; i < itemsCount; i++ {
		copy(s7Item, s7MultiReadItemTelegram)
		s7Item[3] = byte(dataItems[i].WordLen)
		binary.BigEndian.PutUint16(s7Item[4:], uint16(dataItems[i].Amount))
		if dataItems[i].Area == s7areadb {
			binary.BigEndian.PutUint16(s7Item[6:], uint16(dataItems[i].DBNumber))
		}
		s7Item[8] = byte(dataItems[i].Area)

		// Adjusts the offset
		var addr int
		if dataItems[i].WordLen == s7wlcounter || dataItems[i].WordLen == s7wltimer {
			addr = dataItems[i].Start
		} else if dataItems[i].WordLen == s7wlbit {
			addr = dataItems[i].Start << 3
			addr += dataItems[i].Bit // Add Bit addr
		} else {
			addr = dataItems[i].Start * 8
		}

		// Build the offset
		s7Item[11] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7Item[10] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7Item[9] = byte(addr & 0x0FF)
		//now expand array then put item into
		s7Multi = append(s7Multi, s7Item...)
		offset += len(s7Item)
	}
	tt, _ := interface{}(mb.transporter).(*TCPClientHandler)
	if offset > tt.PDULength {
		err = fmt.Errorf(ErrorText(errCliSizeOverPDU))
		return
	}
	binary.BigEndian.PutUint16(s7Multi[2:], uint16(offset)) // Whole size
	request := NewProtocolDataUnit(s7Multi)
	//send
	response, err := mb.send(&request)
	if err != nil {
		return
	}
	// Check ISO Length
	resLength := len(response.Data)
	if resLength < 22 {
		err = fmt.Errorf(ErrorText(errIsoInvalidPDU)) // PDU too Small
		return
	}
	// Check Global Operation Result
	cpuErr := CPUError(uint(binary.BigEndian.Uint16(response.Data[17:])))
	if cpuErr != 0 {
		err = fmt.Errorf(ErrorText(cpuErr))
		return
	}
	// Get true ItemsCount
	itemsRead := int(response.Data[20])
	s7ItemRead := make([]byte, 1024)
	if itemsRead != itemsCount || itemsRead > 20 { //max var
		err = fmt.Errorf(ErrorText(errCliInvalidPlcAnswer))
		return
	}
	// Get Data
	offset = 21
	for i := 0; i < itemsCount; i++ {
		// Get the Item
		copy(s7ItemRead[0:resLength-offset], response.Data[offset:resLength])
		if s7ItemRead[0] == 255 {
			itemSize := int(binary.BigEndian.Uint16(s7ItemRead[2:]))
			item1 := s7ItemRead[1]
			if item1 != tsResOctet && item1 != tsResReal && item1 != tsResBit {
				itemSize = itemSize >> 3
			}
			copy(dataItems[i].Data[0:], s7ItemRead[4:4+itemSize])
			dataItems[i].Error = ""
			if itemSize%2 != 0 {
				itemSize++ // Odd size are rounded
			}
			offset = offset + 4 + itemSize
		} else {
			dataItems[i].Error = ErrorText(CPUError(uint(s7ItemRead[0])))
			offset += 4 // Skip the Item header
		}
	}

	return

}
