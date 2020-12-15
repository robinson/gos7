package gos7

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	bias int64 = 621355968000000000 // "decimicros" between 0001-01-01 00:00:00 and 1970-01-01 00:00:00
)

//Helper the helper to get/set value from/to byte array with difference types
type Helper struct{}

//SetValueAt set a value at a position of a byte array,
//which based on builtin function: https://golang.org/pkg/encoding/binary/#Read
func (s7 *Helper) SetValueAt(buffer []byte, pos int, data interface{}) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	copy(buffer[pos:], buf.Bytes())
}

//GetValueAt set a value at a position of a byte array,
// which based on  builtin function: https://golang.org/pkg/encoding/binary/#Write
func (s7 *Helper) GetValueAt(buffer []byte, pos int, value interface{}) {
	buf := bytes.NewReader(buffer[pos:])
	if err := binary.Read(buf, binary.BigEndian, value); err != nil {
		fmt.Println("binary.Read failed:", err)
	}
}

//GetRealAt 32 bit floating point number (S7 Real) (Range of float32)
func (s7 *Helper) GetRealAt(buffer []byte, pos int) float32 {
	var value uint32
	s7.GetValueAt(buffer, pos, &value)
	float := math.Float32frombits(value)
	return float
}

//SetRealAt 32 bit floating point number (S7 Real) (Range of float32)
func (s7 *Helper) SetRealAt(buffer []byte, pos int, value float32) {
	s7.SetValueAt(buffer, pos, math.Float32bits(value))
}

//GetLRealAt 64 bit floating point number (S7 LReal) (Range of float64)
func (s7 *Helper) GetLRealAt(buffer []byte, pos int) float64 {
	var value uint64
	s7.GetValueAt(buffer, pos, &value)
	float := math.Float64frombits(value)
	return float
}

//SetLRealAt 64 bit floating point number (S7 LReal) (Range of float64)
func (s7 *Helper) SetLRealAt(Buffer []byte, Pos int, Value float64) {
	s7.SetValueAt(Buffer, Pos, math.Float64bits(Value))
}

//GetDateTimeAt DateTime (S7 DATE_AND_TIME)
func (s7 *Helper) GetDateTimeAt(Buffer []byte, Pos int) time.Time {
	var Year, Month, Day, Hour, Min, Sec, MSec int
	Year = decodeBcd(Buffer[Pos])
	if Year < 90 {
		Year = Year + 2000
	} else {
		Year += 1900
	}
	Month = decodeBcd(Buffer[Pos+1])
	Day = decodeBcd(Buffer[Pos+2])
	Hour = decodeBcd(Buffer[Pos+3])
	Min = decodeBcd(Buffer[Pos+4])
	Sec = decodeBcd(Buffer[Pos+5])
	MSec = decodeBcd(Buffer[Pos+6])*10 + decodeBcd(Buffer[Pos+7]>>4)
	return time.Date(Year, time.Month(Month), Day, Hour, Min, Sec, MSec*1000000, time.UTC)
}

//Binary-coded decimal https://en.wikipedia.org/wiki/Binary-coded_decimal
func decodeBcd(b byte) int {
	return int(((b >> 4) * 10) + (b & 0x0F))
}

func encodeBcd(value int) byte {
	return byte(((value / 10) << 4) | (value % 10))
}

//SetDateTimeAt DateTime (S7 DATE_AND_TIME)
func (s7 *Helper) SetDateTimeAt(buffer []byte, pos int, value time.Time) {
	y := value.Year()
	m := int(value.Month())
	d := value.Day()
	h := value.Hour()
	mi := value.Minute()
	s := value.Second()
	if y >= 2000 {
		y -= 2000
	} else {
		y -= 1900
	}
	buffer[pos] = encodeBcd(y)
	buffer[pos+1] = encodeBcd(m)
	buffer[pos+2] = encodeBcd(d)
	buffer[pos+3] = encodeBcd(h)
	buffer[pos+4] = encodeBcd(mi)
	buffer[pos+5] = encodeBcd(s)
	buffer[pos+6] = encodeBcd(value.Nanosecond() / 1000000 / 10)
	buffer[pos+7] = (encodeBcd(value.Nanosecond()/1000000%10) << 4) | encodeBcd(int(value.Weekday()))
}

//GetDateAt DATE (S7 DATE)
func (s7 *Helper) GetDateAt(buffer []byte, pos int) time.Time {
	initDate := time.Date(1990, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	var days int16
	s7.GetValueAt(buffer, pos, &days)
	return initDate.AddDate(0, 0, int(days))
}

//SetDateAt DATE (S7 DATE)
func (s7 *Helper) SetDateAt(buffer []byte, pos int, value time.Time) {
	initDate := time.Date(1990, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	hours := value.Sub(initDate).Hours()
	days := int16(hours / 24)
	s7.SetValueAt(buffer, pos, days)
}

//GetTODAt TOD (S7 TIME_OF_DAY)
func (s7 *Helper) GetTODAt(buffer []byte, pos int) time.Time {
	var ms int32
	s7.GetValueAt(buffer, 0, &ms)
	return time.Date(1970, time.Month(1), 1, 0, 0, 0, int(ms)*1000000, time.UTC)
}

//SetTODAt TOD (S7 TIME_OF_DAY)
func (s7 *Helper) SetTODAt(buffer []byte, pos int, value time.Time) {
	v := int32((value.Hour()*3600 + value.Minute()*60 + value.Second()) * 1000)
	s7.SetValueAt(buffer, pos, v)
}

//GetLTODAt LTOD (S7 1500 LONG TIME_OF_DAY)
func (s7 *Helper) GetLTODAt(Buffer []byte, Pos int) time.Time {
	//S71500 Tick = 1 ns
	var nano int64
	s7.GetValueAt(Buffer, Pos, &nano)
	return time.Date(1970, time.Month(1), 1, 0, 0, 0, int(nano), time.UTC)
}

//SetLTODAt LTOD (S7 1500 LONG TIME_OF_DAY)
func (s7 *Helper) SetLTODAt(buffer []byte, pos int, value time.Time) {
	v := int64((value.Hour()*3600 + value.Minute()*60 + value.Second()) * 1000000000)
	s7.SetValueAt(buffer, pos, v)
}

//GetLDTAt LDT (S7 1500 Long Date and Time)
func (s7 *Helper) GetLDTAt(buffer []byte, pos int) time.Time {
	var nano int64
	s7.GetValueAt(buffer, pos, &nano)
	return time.Date(1970, time.Month(1), 1, 0, 0, 0, int(nano), time.UTC)
}

//SetLDTAt LDT (S7 1500 Long Date and Time)
func (s7 *Helper) SetLDTAt(buffer []byte, pos int, value time.Time) {
	s7.SetValueAt(buffer, pos, value.UnixNano())
}

//GetDTLAt DTL (S71200/1500 Date and Time)
func (s7 *Helper) GetDTLAt(buffer []byte, pos int) time.Time {
	var year uint16
	var nanos int32
	s7.GetValueAt(buffer, pos+0, &year)
	s7.GetValueAt(buffer, pos+8, &nanos)
	return time.Date(int(year), time.Month(int(buffer[pos+2])), int(buffer[pos+3]), int(buffer[pos+5]), int(buffer[pos+6]), int(buffer[pos+7]), int(nanos), time.UTC)
}

//SetDTLAt DTL (S71200/1500 Date and Time)
func (s7 *Helper) SetDTLAt(buffer []byte, pos int, value time.Time) []byte {
	year := uint16(value.Year())
	s7.SetValueAt(buffer, pos, year)
	buffer[pos+2] = byte(value.Month())
	buffer[pos+3] = byte(value.Day())
	buffer[pos+4] = byte(value.Weekday())
	buffer[pos+5] = byte(value.Hour())
	buffer[pos+6] = byte(value.Minute())
	buffer[pos+7] = byte(value.Second())
	nanos := int32(value.Nanosecond())
	s7.SetValueAt(buffer, pos+8, nanos)
	return buffer
}

// Get S5Time
func (s7 *Helper) GetS5TimeAt(buffer []byte, pos int) time.Duration {
	t := decodeBcd(buffer[pos+0]&0b00001111)*100 + decodeBcd(buffer[pos+1])
	switch buffer[pos+0] & 0b00110000 {
	case 0b00000000:
		t *= 10
	case 0b00010000:
		t *= 100
	case 0b00100000:
		t *= 1000
	case 0b00110000:
		t *= 10000
	}
	d, _ := time.ParseDuration(fmt.Sprintf("%dms", t))
	return d
}

//SetS5TimeAt Set S5Time
func (s7 *Helper) SetS5TimeAt(buffer []byte, pos int, value time.Duration) []byte {
	ms := value.Milliseconds()
	switch {
	case ms < 9990:
		buffer[pos+1] = encodeBcd(int(ms) / 10 % 100)
		buffer[pos+0] = encodeBcd(int(ms)/10/100) &^ 0b11110000
	case ms > 100 && ms < 99900:
		buffer[pos+1] = encodeBcd(int(ms) / 100 % 100)
		buffer[pos+0] = encodeBcd(int(ms)/100/100)&^0b11100000 | 0b00010000
	case ms > 1000 && ms < 999000:
		buffer[pos+1] = encodeBcd(int(ms) / 1000 % 100)
		buffer[pos+0] = encodeBcd(int(ms)/1000/100)&^0b11010000 | 0b00100000
	case ms > 10000 && ms < 9990000:
		buffer[pos+1] = encodeBcd(int(ms) / 10000 % 100)
		buffer[pos+0] = encodeBcd(int(ms)/10000/100)&^0b11000000 | 0b00110000
	}
	return buffer
}

//SetStringAt Set String (S7 String)
func (s7 *Helper) SetStringAt(buffer []byte, pos int, maxLen int, value string) []byte {
	buffer[pos] = byte(maxLen)
	var byteLen int
	if maxLen < len(value) {
		byteLen = maxLen
	} else {
		byteLen = len(value)
	}
	buffer[pos+1] = byte(byteLen)
	copy(buffer[pos+2:], []byte(value)[:byteLen])
	return buffer
}

//GetStringAt Get String
func (s7 *Helper) GetStringAt(buffer []byte, pos int) string {
	l := uint8(buffer[pos+1])
	return string(buffer[pos+2 : pos+2+int(l)])
}

//SetWStringAt Set String (WString)
func (s7 *Helper) SetWStringAt(buffer []byte, pos int, maxLen int, value string) []byte {
	chars := []rune(value)
	var sLen int
	if maxLen < len(value) {
		sLen = maxLen
	} else {
		sLen = len(value)
	}
	s7.SetValueAt(buffer, pos+0, int16(maxLen))
	s7.SetValueAt(buffer, pos+2, int16(sLen))
	for i, c := range chars {
		if i >= sLen {
			return buffer
		}
		s7.SetValueAt(buffer, pos+4+i*2, uint16(c))
	}
	return buffer
}

//GetWStringAt Get WString
func (s7 *Helper) GetWStringAt(buffer []byte, pos int) string {
	var l, max int16
	var i int
	var s string
	s7.GetValueAt(buffer, pos+0, &max)
	s7.GetValueAt(buffer, pos+2, &l)
	bs := buffer[pos+4:]
	for i < int(l) {
		var c uint16
		s7.GetValueAt(bs, 0, &c)
		bs = bs[2:]
		s += fmt.Sprintf("%c", c)
		i++
	}
	return s
}

//GetCharsAt Get Array of char (S7 ARRAY OF CHARS)
func (s7 *Helper) GetCharsAt(buffer []byte, pos int, Size int) string {
	return string(buffer[pos : pos+Size])
}

//SetCharsAt Get Array of char (S7 ARRAY OF CHARS)
func (s7 *Helper) SetCharsAt(buffer []byte, pos int, value string) {
	buffer = append(buffer[:pos], append([]byte(value), buffer[pos:]...)...)

}

//GetCounter Get S7 Counter
func (s7 *Helper) GetCounter(value uint16) int {
	return int(decodeBcd(byte(value))*100 + decodeBcd(byte(value>>8)))
}

//GetCounterAt Get S7 Counter at a index
func (s7 *Helper) GetCounterAt(buffer []uint16, index int) int {
	return s7.GetCounter(buffer[index])
}

//ToCounter convert value to s7
func (s7 *Helper) ToCounter(value int) uint16 {
	return uint16(encodeBcd(value/100) + encodeBcd(value%100<<8))
}

//SetCounterAt set a counter at a postion
func (s7 *Helper) SetCounterAt(buffer []uint16, pos int, value int) []uint16 {
	buffer[pos] = s7.ToCounter(value)
	return buffer
}

// SetBoolAt sets a boolean (bit) within a byte at bit position
// without changing the other bits
// it returns the resulted byte
func (s7 *Helper) SetBoolAt(b byte, bitPos uint, data bool) byte {
	if data {
		return b | (1 << bitPos)
	}
	return b &^ (1 << bitPos)
}

// GetBoolAt gets a boolean (bit) from a byte at position
func (s7 *Helper) GetBoolAt(b byte, pos uint) bool {
	return b&(1<<pos) != 0
}
