package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

// ISO Connection Request telegram (contains also ISO Header and COTP Header)
var isoConnectionRequestTelegram = []byte{
	// TPKT (RFC1006 Header)
	3,  // RFC 1006 ID (3)
	0,  // Reserved, always 0
	0,  // High part of packet lenght (entire frame, payload and TPDU included)
	22, // Low part of packet lenght (entire frame, payload and TPDU included)
	// COTP (ISO 8073 Header)
	17,  // PDU Size Length
	224, // CR - Connection Request ID
	0,   // Dst Reference HI
	0,   // Dst Reference LO
	0,   // Src Reference HI
	1,   // Src Reference LO
	0,   // Class + Options Flags
	192, // PDU Max Length ID
	1,   // PDU Max Length HI
	10,  // PDU Max Length LO
	193, // Src TSAP Identifier
	2,   // Src TSAP Length (2 bytes)
	1,   // Src TSAP HI (will be overwritten)
	0,   // Src TSAP LO (will be overwritten)
	194, // Dst TSAP Identifier
	2,   // Dst TSAP Length (2 bytes)
	1,   // Dst TSAP HI (will be overwritten)
	2}   // Dst TSAP LO (will be overwritten)

// TPKT + ISO COTP Header (Connection Oriented Transport Protocol)
var tpktISOTelegram = []byte{ // 7 bytes
	3, 0,
	0, 31, // Telegram Length (Data Size + 31 or 35)
	2, 240, 128} // COTP (see above for info)
// S7 PDU Negotiation Telegram (contains also ISO Header and COTP Header)
var s7PDUNegogiationTelegram = []byte{
	3, 0, 0, 25,
	2, 240, 128, // TPKT + COTP (see above for info)
	50, 1, 0, 0, 4, 0, 0, 8, 0, 0, 240, 0, 0, 1, 0, 1,
	0, 30} // PDU Length Requested = HI-LO Here Default 480 bytes

// S7 Read/Write Request Header (contains also ISO Header and COTP Header)
var s7ReadWriteTelegram = []byte{ // 31-35 bytes
	3, 0,
	0, 31, // Telegram Length (Data Size + 31 or 35)
	2, 240, 128, // COTP (see above for info)
	50,   // S7 Protocol ID
	1,    // Job Type
	0, 0, // Redundancy identification
	5, 0, // PDU Reference //lth this use for request S7 packet id
	0, 14, // Parameters Length
	0, 0, // Data Length = Size(bytes) + 4
	4,              // Function 4 Read Var, 5 Write Var
	1,              // Items count
	18,             // Var spec.
	10,             // Length of remaining bytes
	16,             // Syntax ID
	byte(s7wlbyte), // Transport Size idx=22
	0, 0,           // Num Elements
	0, 0, // DB Number (if any, else 0)
	132,     // Area Type
	0, 0, 0, // Area Offset
	// WR area
	0,    // Reserved
	4,    // Transport size
	0, 0} // Data Length * 8 (if not bit or timer or counter)

// S7 Variable MultiRead Header
var s7MultiReadHeaderTelegram = []byte{
	3, 0,
	0, 31, // Telegram Length
	2, 240, 128, // COTP (see above for info)
	50,   // S7 Protocol ID
	1,    // Job Type
	0, 0, // Redundancy identification
	5, 0, // PDU Reference
	0, 14, // Parameters Length
	0, 0, // Data Length = Size(bytes) + 4
	4, // Function 4 Read Var, 5 Write Var
	1} // Items count (idx 18)

// S7 Variable MultiRead Item
var s7MultiReadItemTelegram = []byte{
	18,             // Var spec.
	10,             // Length of remaining bytes
	16,             // Syntax ID
	byte(s7wlbyte), // Transport Size idx=3
	0, 0,           // Num Elements
	0, 0, // DB Number (if any, else 0)
	132,     // Area Type
	0, 0, 0} // Area Offset

// S7 Variable MultiWrite Header
var s7MultiWriteHeaderTelegram = []byte{
	3, 0,
	0, 31, // Telegram Length
	2, 240, 128, // COTP (see above for info)
	50,   // S7 Protocol ID
	1,    // Job Type
	0, 0, // Redundancy identification
	5, 0, // PDU Reference
	0, 14, // Parameters Length (idx 13)
	0, 0, // Data Length = Size(bytes) + 4 (idx 15)
	5, // Function 5 Write Var
	1} // Items count (idx 18)

// S7 Variable MultiWrite Item (Param)
var s7MultiWriteItemTelegram = []byte{
	18,             // Var spec.
	10,             // Length of remaining bytes
	16,             // Syntax ID
	byte(s7wlbyte), // Transport Size idx=3
	0, 0,           // Num Elements
	0, 0, // DB Number (if any, else 0)
	132,     // Area Type
	0, 0, 0} // Area Offset

// SZL First telegram request
var s7SZLFirstTelegram = []byte{
	3, 0, 0, 33, 2, 240, 128, 50, 7, 0, 0,
	5, 0, // Sequence out
	0, 8, 0, 8, 0, 1, 18, 4, 17, 68, 1, 0, 255, 9, 0, 4,
	0, 0, // ID (29)
	0, 0} // Index (31)

// SZL Next telegram request
var s7SZLNextTelegram = []byte{
	3, 0, 0, 33, 2, 240, 128, 50, 7, 0, 0, 6, 0, 0, 12, 0, 4, 0, 1, 18, 8, 18, 68, 1,
	1, // Sequence
	0, 0, 0, 0, 10, 0, 0, 0}

// Get Date/Time request
var s7GetDatetimeTelegram = []byte{
	3, 0, 0, 29, 2, 240, 128, 50, 7, 0, 0, 56, 0, 0, 8, 0, 4, 0, 1, 18, 4, 17, 71, 1, 0, 10, 0, 0, 0}

// Set Date/Time command
var s7SetDatetimeTelegram = []byte{
	3, 0, 0, 39, 2, 240, 128, 50, 7, 0, 0, 137, 3, 0, 8, 0, 14, 0, 1, 18, 4, 17, 71, 2, 0, 255, 9, 0, 10, 0,
	25,   // Hi part of Year (idx=30)
	19,   // Lo part of Year
	18,   // Month
	6,    // Day
	23,   // Hour
	55,   // Min
	19,   // Sec
	0, 1} // ms + Day of week

// S7 Set Session Password
var s7SetPWDTelegram = []byte{
	3, 0, 0, 37, 2, 240, 128, 50, 7, 0, 0, 39, 0, 0, 8, 0, 12, 0, 1, 18, 4, 17, 69, 1, 0, 255, 9, 0,
	8, // 8 Char Encoded Password
	0, 0, 0, 0, 0, 0, 0, 0}

// S7 Clear Session Password
var s7ClearPWDTelegram = []byte{
	3, 0, 0, 29, 2, 240, 128, 50, 7, 0, 0, 41, 0, 0, 8, 0, 4, 0, 1, 18, 4, 17, 69, 2, 0, 10, 0, 0, 0}

// S7 STOP request
var s7StopTelegram = []byte{
	3, 0, 0, 33, 2, 240, 128, 50, 1, 0, 0, 14, 0, 0, 16, 0, 0, 41, 0, 0,
	0, 0, 0, 9, 80, 95, 80, 82, 79, 71, 82, 65, 77}

// S7 HOT Start request
var s7HotStartTelegram = []byte{
	3, 0, 0, 37, 2, 240, 128, 50, 1, 0, 0, 12, 0, 0, 20, 0, 0, 40, 0, 0,
	0, 0, 0, 0, 253, 0, 0, 9, 80, 95, 80, 82, 79, 71, 82, 65, 77}

// S7 COLD Start request
var s7ColdStartTelegram = []byte{
	3, 0, 0, 39, 2, 240, 128, 50, 1, 0, 0, 15, 0, 0, 22, 0, 0, 40, 0, 0,
	0, 0, 0, 0, 253, 0, 2, 67, 32, 9, 80, 95, 80, 82, 79, 71, 82, 65, 77}

const (
	pduStart          = 0x28 // CPU start
	pduStop           = 0x29 // CPU stop
	pduAlreadyStarted = 0x02 // CPU already in run mode
	pduAlreadyStopped = 0x07 // CPU already in stop mode
)

// S7 Get PLC Status
var s7GetStatusTelegram = []byte{
	3, 0, 0, 33, 2, 240, 128, 50, 7, 0, 0, 44, 0, 0, 8, 0, 8, 0,
	1, 18, 4, 17, 68, 1, 0, 255, 9, 0, 4, 4, 36, 0, 0}

// S7 Get Block Info Request Header (contains also ISO Header and COTP Header)
var s7BlockInfoTelegram = []byte{
	3, 0, 0, 37, 2, 240, 128, 50, 7, 0, 0, 5, 0, 0, 8, 0, 12,
	0, 1, 18, 4, 17, 67, 3, 0, 255, 9, 0, 8, 48,
	65,                 // Block Type
	48, 48, 48, 48, 48, // ASCII Block Number
	65}

//s7 pg block list telegram, require type to the end
var s7PGBlockListTelegram = []byte{
	3, 0, 0, 31, 2, 240, 128, 50, 7, 0, 0, 5, 0, 0, 8, 0, 6, 0, 1, 18, 4, 17, 67, 2, 0, 255, 9, 0, 2, 48}

var s7PGBlockDeleteTelegram = []byte{
	50, 1, 0, 0, 107, 0, 0, 26, 0, 0, 40, 0, 0, 0, 0, 0, 0, 253, 0, 10, 1, 0, 48,
	0,             //block type, should replace
	0, 0, 0, 0, 0, //block number, should replace
	66, 5, // should add byte
	95, 68, 69, 76, 69} //bytes of string "_DELE"
