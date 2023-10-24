# gos7
Implementation of Siemens S7 protocol in golang

Overview
-------------------
For years, numerous drivers/connectors, available in both commercial and open source domains, have supported the connection to S7 family PLC devices. GoS7 fills the gaps in the S7 protocol, implementing it with pure Go (also known as golang). There is a strong belief that low-level communication should be implemented with a low-level programming language that is close to binary and memory.

The minimum supported Go version is 1.13.

Functions
-------------------
AG:
*   Read/Write Data Block (DB) (tested)
*   Read/Write Merkers(MB) (tested)
*   Read/Write IPI (EB) (tested)
*   Read/Write IPU (AB) (tested)
*   Read/Write Timer (TM)  (tested)
*   Read/Write Counter (CT) (tested)
*   Multiple Read/Write Area (tested)
*   Get Block Info (tested)

PG:
*   Hot start/Cold start / Stop PLC
*   Get CPU of PLC status (tested)
*   List available blocks in PLC (tested)
*   Set/Clear password for session
*   Get CPU protection and CPU Order code
*   Get CPU/CP Information (tested)
*   Read/Write clock for the PLC
Helpers:
*   Get/set value for a byte array for types: value(bit/int/word/dword/uint...), real, time, counter

Supported communication
-----------------
*   TCP
*   Serial (PPI, MPI) (under construction)

How to:
----------
following is a simple usage to connect with PLC via TCP
```go
const (
	tcpDevice = "127.0.0.1"
	rack      = 0
	slot      = 2
)
// TCPClient
handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
handler.Timeout = 200 * time.Second
handler.IdleTimeout = 200 * time.Second
handler.Logger = log.New(os.Stdout, "tcp: ", log.LstdFlags)
// Connect manually so that multiple requests are handled in one connection session
handler.Connect()
defer handler.Close()
//init client
client := gos7.NewClient(handler)
address := 2710
start := 8
size := 2
buffer := make([]byte, 255)
value := 100
//AGWriteDB to address DB2710 with value 100, start from position 8 with size = 2 (for an integer)
var helper gos7.Helper
helper.SetValueAt(buffer, 0, value)  
err := client.AGWriteDB(address, start, size, buffer)
buf := make([]byte, 255)
//AGReadDB to address DB2710, start from position 8 with size = 2
err := client.AGReadDB(address, start, size, buf)
var s7 gos7.Helper
var result uint16
s7.GetValueAt(buf, 0, &result)	 
  
```
References
----------
- libnodave http://libnodave.sourceforge.net/
- snap7 http://snap7.sourceforge.net/ 
- tarm serial library https://github.com/tarm/serial
- Simatic Open TCP/IP Communication via Industrial Ethernet from Siemens(doku)
- SIMATIC NET FDL-Programmierschnittstelle (doku)
- Elementary Data Types from Siemens (doku)

Simatic, Simatic S5, Simatic S7, S7-200, S7-300, S7-400, S7-1200, S7-1500 are registered Trademarks of Siemens

License
----------
https://opensource.org/licenses/BSD-3-Clause

Copyright (c) 2018, robinson
