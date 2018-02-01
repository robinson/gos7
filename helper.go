package gos7

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
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
	MSec = (decodeBcd(Buffer[Pos+6]) * 10) + (decodeBcd(Buffer[Pos+7]) / 10)
	return time.Date(Year, time.Month(Month), Day, Hour, Min, Sec, MSec, time.UTC)
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
	dow := int(value.Weekday()) + 1
	// msh = First two digits of miliseconds
	msh := int(int64(value.UnixNano()/1000000) / 10)
	// msl = Last digit of miliseconds
	msl := int(int64(value.UnixNano()/1000000) % 10)
	if y > 1999 {
		y -= 2000
	}
	buffer[pos] = encodeBcd(y)
	buffer[pos+1] = encodeBcd(m)
	buffer[pos+2] = encodeBcd(d)
	buffer[pos+3] = encodeBcd(h)
	buffer[pos+4] = encodeBcd(mi)
	buffer[pos+5] = encodeBcd(s)
	buffer[pos+6] = encodeBcd(msh)
	buffer[pos+7] = encodeBcd(msl*10 + dow)
}

//GetDateAt DATE (S7 DATE)
func (s7 *Helper) GetDateAt(buffer []byte, pos int) time.Time {
	initDate := time.Date(1900, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	var year int16
	s7.GetValueAt(buffer, pos, &year)
	return initDate.AddDate(0, 0, int(year))
}

//SetDateAt DATE (S7 DATE)
func (s7 *Helper) SetDateAt(buffer []byte, pos int, value time.Time) {
	initDate := time.Date(1900, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	s7.SetValueAt(buffer, pos, int16(value.YearDay()-initDate.YearDay()))
}

//GetTODAt TOD (S7 TIME_OF_DAY)
func (s7 *Helper) GetTODAt(buffer []byte, pos int) time.Time {
	var nano int32
	s7.GetValueAt(buffer, pos, &nano)
	return time.Date(0001, 1, 1, 0, 0, int(nano/1000), 0, time.UTC)
}

//SetTODAt TOD (S7 TIME_OF_DAY)
func (s7 *Helper) SetTODAt(buffer []byte, pos int, value time.Time) {
	s7.SetValueAt(buffer, pos, int32(value.Nanosecond()/1000000))
}

//GetLTODAt LTOD (S7 1500 LONG TIME_OF_DAY)
func (s7 *Helper) GetLTODAt(Buffer []byte, Pos int) time.Time {
	//S71500 Tick = 1 ns
	var nano int64
	s7.GetValueAt(Buffer, Pos, &nano)
	return time.Date(0, 0, 0, 0, 0, 0, int(nano), time.UTC)
}

//SetLTODAt LTOD (S7 1500 LONG TIME_OF_DAY)
func (s7 *Helper) SetLTODAt(buffer []byte, pos int, value time.Time) {
	s7.SetValueAt(buffer, pos, int64(value.Nanosecond()))
}

//GetLDTAt LDT (S7 1500 Long Date and Time)
func (s7 *Helper) GetLDTAt(buffer []byte, pos int) time.Time {
	var nano int64
	s7.GetValueAt(buffer, pos, &nano)
	return time.Date(0, 0, 0, 0, 0, 0, int(nano+bias), time.UTC)
}

//SetLDTAt LDT (S7 1500 Long Date and Time)
func (s7 *Helper) SetLDTAt(buffer []byte, pos int, value time.Time) {
	s7.SetValueAt(buffer, pos, int64(value.Nanosecond())-bias)
}

//GetDTLAt DTL (S71200/1500 Date and Time)
func (s7 *Helper) GetDTLAt(buffer []byte, pos int) time.Time {
	Year := int(buffer[pos])*256 + int(buffer[pos+1])
	Month := int(buffer[pos+2])
	Day := int(buffer[pos+3])
	Hour := int(buffer[pos+5])
	Min := int(buffer[pos+6])
	Sec := int(buffer[pos+7])
	var nsec int
	s7.GetValueAt(buffer, pos, &nsec)
	return time.Date(Year, time.Month(Month), Day, Hour, Min, Sec, nsec, time.UTC)
}

//SetDTLAt DTL (S71200/1500 Date and Time)
func (s7 *Helper) SetDTLAt(buffer []byte, pos int, value time.Time) []byte {
	Year := []byte(strconv.Itoa(value.Year()))
	buffer[pos] = Year[1]
	buffer[pos+1] = Year[0]
	buffer[pos+2] = byte(value.Month())
	buffer[pos+3] = byte(value.Day())
	buffer[pos+4] = byte(int(value.Weekday()) + 1)
	buffer[pos+5] = byte(value.Hour())
	buffer[pos+6] = byte(value.Minute())
	buffer[pos+7] = byte(value.Second())
	buffer[pos+7] = byte(value.Nanosecond())
	return buffer
}

//SetStringAt Set String (S7 String)
func (s7 *Helper) SetStringAt(buffer []byte, pos int, maxLen int, value string) []byte {
	buffer[pos] = byte(maxLen)
	buffer[pos+1] = byte(len(value))
	buffer = append(buffer[:pos+2], append([]byte(value), buffer[pos+2:]...)...)
	return buffer
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
