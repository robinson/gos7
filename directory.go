package gos7

import (
	"fmt"
)

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
const (
	// Block type byte
	blockOB  = 56
	blockDB  = 65
	blockSDB = 66
	blockFC  = 67
	blockSFC = 68
	blockFB  = 69
	blockSFB = 70
)

//S7BlocksList Block List
type S7BlocksList struct {
	OBList  []int
	FBList  []int
	FCList  []int
	SFBList []int
	SFCList []int
	DBList  []int
	SDBList []int
}

//implement list block
func (mb *client) PGListBlocks() (list S7BlocksList, err error) {
	list.OBList, err = mb.pgBlockList(blockOB)
	//debug
	fmt.Printf("%v", list.DBList)
	list.DBList, err = mb.pgBlockList(blockDB)
	list.FCList, err = mb.pgBlockList(blockFC)
	list.OBList, err = mb.pgBlockList(blockOB)
	list.FBList, err = mb.pgBlockList(blockFB)
	list.SDBList, err = mb.pgBlockList(blockSDB)
	list.SFBList, err = mb.pgBlockList(blockSFB)
	list.SFCList, err = mb.pgBlockList(blockSFC)
	return
}

func (mb *client) pgBlockList(blockType byte) (arr []int, err error) {
	bl := make([]byte, len(s7PGBlockListTelegram))
	copy(bl, s7PGBlockListTelegram)
	bl = append(bl, make([]byte, 1)...)
	switch blockType {
	case blockDB:
		bl[len(bl)-1] = blockDB
	case blockOB:
		bl[len(bl)-1] = blockOB
	case blockSDB:
		bl[len(bl)-1] = blockSDB
	case blockFC:
		bl[len(bl)-1] = blockFC
	case blockSFC:
		bl[len(bl)-1] = blockSFC
	case blockFB:
		bl[len(bl)-1] = blockFB
	case blockSFB:
		bl[len(bl)-1] = blockSFB
	default:
		return
	}
	request := NewProtocolDataUnit(bl)
	//send
	response, err := mb.send(&request)
	if err == nil {
		res := make([]byte, len(response.Data)-33) //remove first 26 byte function and 7 byte header
		copy(res, response.Data[33:len(response.Data)])
		arr = dataToBlocks(res)
	}
	return
}
func dataToBlocks(data []byte) []int {
	arr := make([]int, len(data)/4)
	for i := 0; i <= len(data)/4-1; i++ {
		arr[i] = int(data[i*4])*256 + int(data[i*4+1])
	}
	return arr
}
