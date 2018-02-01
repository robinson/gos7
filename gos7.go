package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"fmt"
	"strconv"
)

// S7Error implements error interface.
type S7Error struct {
	High byte
	Low  byte
}

// Packager specifies the communication layer.
type Packager interface {
	//reserve for future use
	Verify(request []byte, response []byte) (err error)
}

// ProtocolDataUnit (PDU) is independent of underlying communication layers.
type ProtocolDataUnit struct {
	Data []byte
}

//NewProtocolDataUnit ProtocolDataUnit Constructor
func NewProtocolDataUnit(data []byte) ProtocolDataUnit {
	pdu := ProtocolDataUnit{Data: data}
	return pdu
}

// Transporter specifies the transport layer.
type Transporter interface {
	Send(request []byte) (response []byte, err error)
}

// Error converts known s7 exception code to error message.
func (e *S7Error) Error() string {
	/* CPU tells there is no peripheral at address */
	errMsg := int(e.High)*256 + int(e.Low)
	message := "UNKNOWN ERROR: " + strconv.Itoa(errMsg)
	switch errMsg {
	case 65487:
		message = "API function called with an invalid parameter"
	case 65535:
		message = "timeout, check RS232 interface"
	case 56321:
		message = "maybe invalid BCD code or Invalid time format"
	case 61185:
		message = "wrong ID2, cyclic job handle"
	case 54278:
		message = "information doesn’t exist"
	case 54281:
		message = "diagnosis: DP Error"
	case 55298:
		message = "this job does not exist"
	case 53824:
		message = "coordination rules were violated"
	case 53825:
		message = "protection level too low"
	case 53826:
		message = "protection violation while processing F-blocks; F-blocks can only be processed after password input"
	case 54273:
		message = "invalid SSL ID"
	case 54274:
		message = "invalid SSL index"
	case 53409:
		message = "Step7: function is not allowed in the current protection level."
	case 53761:
		message = "syntax error: block name"
	case 53762:
		message = "syntax error: function parameter"
	case 53763:
		message = "syntax error: block type"
	case 53764:
		message = "no linked data block in CPU"
	case 53765:
		message = "object already exists"
	case 53766:
		message = "object already exists"
	case 53767:
		message = "data block in EPROM"
	case 53769:
		message = "block doesn’t exist"
	case 53774:
		message = "no block available"
	case 53776:
		message = "block number too large"
	case 34048:
		message = "wrong PDU (response data) size"
	case 34562:
		message = "Not address"
	case 53250:
		message = "Step7: variant of command is illegal."
	case 53252:
		message = "Step7: status for this command is illegal."
	case 33537:
		message = "not enough memory on CPU"
	case 33794:
		message = "maybe CPU already in RUN or already in STOP"
	case 33796:
		message = "serious error"
	case 32768:
		message = "interface Is busy"
	case 32769:
		message = "not permitted in this mode"
	case 33025:
		message = "hardware error"
	case 33027:
		message = "access to object not permitted"
	case 33028:
		message = "Not context"
	case 33029:
		message = "address invalid. This may be due to a memory address that is not valid for the PLC"
	case 33030:
		message = "data type not supported"
	case 33031:
		message = "data type not consistent"
	case 33034:
		message = "object doesn’t exist. This may be due to a data block that doesn’t exist in the PLC"
	case 800:
		message = "hardware error"
	case 897:
		message = "hardware error"
	case 16385:
		message = "communication link unknown"
	case 16386:
		message = "communication link not available"
	case 16387:
		message = "MPI communication in progress"
	case 16388:
		message = "MPI connection down; this may be due to an invalid MPI address (local or remote ID) or the PLC is not communicating on the MPI network"
	case 512:
		message = "unknown error"
	case 513:
		message = "wrong interface specified"
	case 514:
		message = "too many interfaces"
	case 515:
		message = "interface already initialized"
	case 516:
		message = "interface already initialized with another connection"
	case 517:
		message = "interface not initialized; this may be due to an invalid MPI address (local or remote ID) or the PLC is not communicating on the MPI network"
	case 518:
		message = "can’t set handle"
	case 519:
		message = "data segment isn’t locked"
	case 521:
		message = "data field incorrect"
	case 770:
		message = "block size is too small"
	case 771:
		message = "block boundary exceeded"
	case 787:
		message = "wrong MPI baud rate selected"
	case 788:
		message = "highest MPI address is wrong"
	case 789:
		message = "address already exists"
	case 794:
		message = "not connected to MPI network"
	case 795:
		message = "-"
	case 1:
		/* CPU tells there is no peripheral at address */
		message = "No data from I/O module" //"hardware fault"
	case 3:
		/* means a a piece of data is not available in the CPU, e.g. */
		/* when trying to read a non existing DB or bit bloc of length<>1 */
		/* This code seems to be specific to 200 family. */
		message = "object access not allowed: occurs when access to timer and counter data type is set to signed integer and not BCD"
	case 4:
		message = "Not context"
	case 5:
		/* means the data address is beyond the CPUs address range */
		message = "the desired address is beyond limit for this PLC" //"address out of range: occurs when requesting an address within a data block that does not exist or is out of range"
	case 6:
		/* CPU tells it does not support to read a bit block with a */
		/* length other than 1 bit. */
		message = "the CPU does not support reading a bit block of length<>1" //"address out of range"
	case 7:
		/* means the write data size doesn't fit item size */
		message = "Write data size error" //"write data size mismatch"
	case 10:
		/* means a a piece of data is not available in the CPU, e.g. */
		/* when trying to read a non existing DB */
		message = "the desired item is not available in the PLC" //"object does not exist: occurs when trying to request a data block that does not exist"
	case 257:
		message = "communication link not available"
	case 266:
		message = "negative acknowledge / time out error"
	case 268:
		message = "data does not exist or is locked"
	case -123:
		/* PDU is not understood by libnodave */
		message = "cannot evaluate the received PDU"
	case -124:
		message = "the PLC returned a packet with no result data"
	case -125:
		message = "the PLC returned an error code not understood by this library"
	case -126:
		message = "this result contains no data"
	case -127:
		message = "cannot work with an undefined result set"
	case -128:
		message = "Unexpected function code in answer"
	case -129:
		message = "PLC responds with an unknown data type"
	case -130:
		message = "No buffer provided"
	case -131:
		message = "Function not supported for S5"
	// case -132:
	// case -133:
	// case -134:
	case -1024:
		message = "Short packet from PLC"
	case -1025:
		message = "Timeout when waiting for PLC response"
	default:
		message = "UNKNOWN ERROR: " + strconv.Itoa(errMsg)
	}
	return fmt.Sprintf("S7: exception (%s)'", message)
}
