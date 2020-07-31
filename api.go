package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"time"
)

//Client interface s7 client
type Client interface {
	/***************start API AG (Automatisationsgerät)***************/
	//Read data blocks from PLC
	AGReadDB(dbNumber int, start int, size int, buffer []byte) (err error)
	//write data blocks into PLC
	AGWriteDB(dbNumber int, start int, size int, buffer []byte) (err error)
	//Read Merkers area from PLC
	AGReadMB(start int, size int, buffer []byte) (err error)
	//Write Merkers from into PLC
	AGWriteMB(start int, size int, buffer []byte) (err error)
	//Read IPI from PLC
	AGReadEB(start int, size int, buffer []byte) (err error)
	//Write IPI into PLC
	AGWriteEB(start int, size int, buffer []byte) (err error)
	//Read IPU from PLC
	AGReadAB(start int, size int, buffer []byte) (err error)
	//Write IPU into PLC
	AGWriteAB(start int, size int, buffer []byte) (err error)
	//Read timer from PLC
	AGReadTM(start int, size int, buffer []byte) (err error)
	//Write timer into PLC
	AGWriteTM(start int, size int, buffer []byte) (err error)
	//Read counter from PLC
	AGReadCT(start int, size int, buffer []byte) (err error)
	//Write counter into PLC
	AGWriteCT(start int, size int, buffer []byte) (err error)
	//multi read area
	AGReadMulti(dataItems []S7DataItem, itemsCount int) (err error)
	//multi write area
	AGWriteMulti(dataItems []S7DataItem, itemsCount int) (err error)
	/*block*/
	DBFill(dbnumber int, fillchar int) error
	DBGet(dbnumber int, usrdata []byte, size int) error
	//general read function with S7 sytax
	Read(variable string, buffer []byte) (value interface{}, err error)
	//Get block  infor in AG area, refer an S7BlockInfor pointer
	GetAgBlockInfo(blocktype int, blocknum int) (info S7BlockInfo, err error)
	/***************end API AG***************/

	/***************start API PG (Programmiergerät)***************/
	/*control*/
	//Hotstart PLC, Puts the CPU in RUN mode performing an HOT START.
	PLCHotStart() error
	//Cold start PLC, change CPU into runmode performing and COLD START
	PLCColdStart() error
	//change CPU to stop mode
	PLCStop() error
	//return CPU status: running/stopped
	PLCGetStatus() (status int, err error)
	/*directory*/
	//list all blocks in PLC, return a Blockslist which contains list of OB, DB, ...
	PGListBlocks() (list S7BlocksList, err error)
	/*security*/
	//set the session password for PLC to meet its security level
	SetSessionPassword(password string) error
	//clear the password set for current session
	ClearSessionPassword() error
	//return the CPU protection level info, refer to: §33.19 of "System Software for S7-300/400 System and Standard Functions"
	//return S7Protection and its properties.
	GetProtection() (protection S7Protection, err error)
	/*system information*/
	//get CPU order code, return S7OrderCode
	GetOrderCode() (info S7OrderCode, err error)
	//get CPU info, return S7CpuInfo and its properties
	GetCPUInfo() (info S7CpuInfo, err error)
	//get CP info, return S7CpInfo and its properties
	GetCPInfo() (info S7CpInfo, err error)
	/*datetime*/
	//read clock on PLC, return a time
	PGClockRead(datetime time.Time) error
	//write clock to PLC with datetime input
	PGClockWrite() (dt time.Time, err error)
	/***************end API AG***************/
}
