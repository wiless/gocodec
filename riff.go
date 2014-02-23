package gocodec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

type Riff struct {
	Header            [4]byte // Fixed "RIFF"
	DataLen           uint32
	FormatStringFixed [8]byte // Fixed "WAVEfmt "
	FormatSizeFixed   uint32  // Fixed value 16
	FormatFixed       uint16  // Fixed value 1
	Channels          uint8   // Number of Channels Mono=1, Stereo=2
	UnknownByte       byte    // Null = 0
	SampleRate        uint32  // 8K, 11K, etc..
	ByteRate          uint32  // Bytes Per Seconds
	FixedValue1       uint16  // Fixed value 2
	SampleWidth       uint8   // Bits Per Sample
	FixedValue2       uint8   // Fixed value 0
	DataString        [4]byte // Fixed value "data"
	PayloadSizeBytes  uint32  // Size of Payload in bytes (excluding 44byte header)
}

func CreateRIFF(sampleRate uint32, bitsPerSample uint8, stereo bool) Riff {
	var r Riff
	/// set all fixed values
	copy(r.Header[0:4], []byte("RIFF"))
	copy(r.FormatStringFixed[0:8], []byte("WAVEfmt"))
	copy(r.DataString[0:4], []byte("data"))
	r.FormatSizeFixed = 16
	r.FormatFixed = 1
	r.UnknownByte = 0
	r.FixedValue1 = 2
	r.FixedValue2 = 0

	/// Dynamic Values
	r.PayloadSizeBytes = 0
	r.DataLen = 0

	/// User Defined
	if stereo {
		r.Channels = 2
	} else {
		r.Channels = 1
	}

	r.SampleRate = sampleRate
	r.SampleWidth = bitsPerSample
	r.ByteRate = r.SampleRate * uint32(r.SampleWidth/8)

	return r
}

func ParseFile(filename string) Riff {
	var tmpriff Riff

	bytearray, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Errorf("Error Parsing %v", err)
		return tmpriff
	}
	tmpriff.Parse(bytearray)
	return tmpriff
}

func (r *Riff) Parse(rawheader []byte) {

	if len(rawheader) < 44 {
		fmt.Errorf("\nUnable to parse %d", len(rawheader))
	} else {
		binary.Read(bytes.NewBuffer(rawheader[0:44]), binary.LittleEndian, r)
	}

}
func (r Riff) String() string {
	// RIFF (little-endian) data, WAVE audio, Microsoft PCM, 16 bit, mono 8000 Hz
	var result string

	result = string(r.Header[0:4])
	if result != "RIFF" {
		return "Unknown"
	}
	result += " LE " + string(r.FormatStringFixed[0:8]) + " , " + strconv.Itoa(int(r.SampleWidth)) + " bits "
	if r.Channels == 1 {
		result += " mono @ " + strconv.Itoa(int(r.SampleRate)) + " Hz"
	} else {
		result += " stereo @ " + strconv.Itoa(int(r.SampleRate)) + " Hz"
	}
	lengthInSeconds := time.Duration(r.Duration()) * time.Second
	result += " Duration : " + lengthInSeconds.String()
	return result
}

func (r Riff) Duration() int {
	return int(r.PayloadSizeBytes / r.ByteRate)
}

func (r Riff) Bytes() []byte {
	// RIFF (little-endian) data, WAVE audio, Microsoft PCM, 16 bit, mono 8000 Hz

	var temp []byte
	bytebuffer := bytes.NewBuffer(temp)
	binary.Write(bytebuffer, binary.LittleEndian, r)

	return bytebuffer.Bytes()
}
