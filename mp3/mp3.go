package mp3

import (
	"fmt"
)

var bitmask [8]int

type Row [5]int

var BitRateMap map[uint8]Row

// See references  http://www.datavoyage.com/mpgscript/mpeghdr.htm
type Mp3Frame struct {
	header           [4]byte
	FrameSync        bool
	AudioVersion     uint8 /// 0 to 3 , MPEG 2.5, Reserved, MPeg2, Mpeg1
	LayerDescription uint8 /// 0 reserved to 3, Layer 3, Layer 2, Layer 1
	ProtectionBit    bool
	BitRateIndex     uint8 /// ranges 0 to 15
	SampleRate       uint8 /// 0 to 3 :
	PaddingBit       bool  ///  Contains frame padding or not
	PrivateBit       bool  /// Freely used
	ChannelMode      uint8 /// 0 to 3, Stereo, Joint Stereo, Dual Channel, Mono
	ModeExtension    uint8 /// 0 to 3, For Joint Stereo
	CopyRight        bool  /// 0 Audio has No copyright, 1 = yes
	Original         bool  /// 0 = Copy of the Original Media
	Emphasis         uint8 /// 0 to 3, 00 - none 1 - 50/15 ms 2 - reserved  3 - CCIT J.17

}

func LoadBitRateTable() {
	bitmask = [8]int{1, 2, 4, 8, 16, 32, 64, 128}

	BitRateMap = make(map[uint8]Row)

	BitRateMap[0] = Row([5]int{0, 0, 0, 0, 0})
	BitRateMap[1] = Row([5]int{32, 32, 32, 32, 8})
	BitRateMap[2] = Row([5]int{64, 48, 40, 48, 16})
	BitRateMap[3] = Row([5]int{96, 56, 48, 56, 24})
	BitRateMap[4] = Row([5]int{128, 64, 56, 64, 32})
	BitRateMap[5] = Row([5]int{160, 80, 64, 80, 40})
	BitRateMap[6] = Row([5]int{192, 96, 80, 96, 48})
	BitRateMap[7] = Row([5]int{224, 112, 96, 112, 56})
	BitRateMap[8] = Row([5]int{256, 128, 112, 128, 64})
	BitRateMap[9] = Row([5]int{288, 160, 128, 144, 80})
	BitRateMap[10] = Row([5]int{320, 192, 160, 160, 96})
	BitRateMap[11] = Row([5]int{352, 224, 192, 176, 112})
	BitRateMap[12] = Row([5]int{384, 256, 224, 192, 128})
	BitRateMap[13] = Row([5]int{416, 320, 256, 224, 144})
	BitRateMap[14] = Row([5]int{448, 384, 320, 256, 160})
	BitRateMap[15] = Row([5]int{-1, -1, -1, -1, -1})
}
func (m Mp3Frame) String() string {
	var result string
	for i := 0; i < 4; i++ {
		result += fmt.Sprintf("%v", m.header[i])
	}
	return fmt.Sprintf("%v", m.header)
}

func maskAndShift(input byte, position int, length int) byte {

	output := 0
	result := 0
	// fmt.Printf("\n Input : %0X, search %d, %d", input, position, length)
	for i := 0; i < length; i++ {
		result = result | bitmask[7-(i+position)]
	}
	offset := uint(7 - (position + length - 1))
	output = (int(input) & result) >> offset
	// fmt.Printf("Mask : %0X  Condition %0X , output is %0X", result, (int(input) & result), output)
	return byte(output)
}

func findSampleRate(indx, aversion uint8) int {

	// bits	MPEG1	MPEG2	MPEG2.5
	var datarow []int
	var colindx int
	if aversion == 3 {
		colindx = 0
	} else if aversion == 2 {
		colindx = 1
	} else if aversion == 0 {
		colindx = 2
	}

	if indx == 0 {
		datarow = []int{44100, 22050, 11025}
		return datarow[colindx]
	}
	if indx == 1 {
		datarow = []int{48000, 24000, 12000}
		return datarow[colindx]
	}
	if indx == 2 {
		datarow = []int{32000, 16000, 8000}
		return datarow[colindx]
	}

	return 0
}
func findBitRate(indx, aversion, layer uint8) int {

	colindx := -1
	row := BitRateMap[indx]
	// bits	V1,L1	V1,L2	V1,L3	V2,L1	V2, L2 & L3
	// 	V1 - MPEG Version 1
	// V2 - MPEG Version 2 and Version 2.5
	var str string
	if aversion == 3 {
		str = "V1"
	} else {
		str = "V2"
	}

	if layer == 3 {
		str += "L1"
	} else if layer == 2 {
		str += "L2"
	} else if layer == 1 {
		str += "L3"
	}

	// L1 - Layer I
	// L2 - Layer II
	// L3 - Layer III

	switch {
	case str == "V1L1":
		colindx = 0
	case str == "V1L2":
		colindx = 1
	case str == "V1L3":
		colindx = 2
	case str == "V2L1":
		colindx = 3
	case str == "V2L2" || str == "V2L3":
		colindx = 4
	}

	if colindx == -1 {
		fmt.Printf("\n Unknown Column String %s", str)
		return 3333
	}
	return row[colindx]
}

func (frame *Mp3Frame) GetBitRate() int {

	return findBitRate(frame.BitRateIndex, frame.AudioVersion, frame.LayerDescription)
}

func (frame *Mp3Frame) GetSampleFreq() int {
	return findSampleRate(frame.SampleRate, frame.AudioVersion)
}

func (frame *Mp3Frame) GetFrameLengthBytes() int {
	// For Layer I files us this formula:
	var FrameLengthInBytes int = -1000
	var str string
	layer := frame.LayerDescription
	if layer == 3 {
		str = "L1"
	} else if layer == 2 {
		str = "L2"
	} else if layer == 1 {
		str = "L3"
	}
	BitRate := frame.GetBitRate() * 1000
	SampleRate := frame.GetSampleFreq()
	Padding := 0
	if frame.PaddingBit {
		Padding = 1
	}
	if SampleRate == 0 {
		// fmt.Printf("\n SamplateFrequench is ZERO Hz reserved ")
		return 0
	}
	if str == "L1" {
		// For Layer I files use this formula:

		FrameLengthInBytes = (12*BitRate/SampleRate + Padding) * 4
	} else {
		// For Layer II & III files use this formula:

		FrameLengthInBytes = 144*BitRate/SampleRate + Padding
	}
	return FrameLengthInBytes
}

/// Pushes a byte to the Frame and validates if its a sync frame
func (frame *Mp3Frame) PushAndValidate(databyte byte) bool {
	temp := append(frame.header[1:4], databyte)
	copy(frame.header[:], temp)
	frame.FrameSync = false
	if frame.header[0] == 0xFF { /// first 8 bits

		output := maskAndShift(frame.header[1], 0, 3) /// next 3 bits
		frame.FrameSync = (output == 7)
		if frame.FrameSync {
			/// Using 2nd Byte

			frame.AudioVersion = maskAndShift(frame.header[1], 3, 2)
			frame.LayerDescription = maskAndShift(frame.header[1], 5, 2)
			frame.ProtectionBit = maskAndShift(frame.header[1], 7, 1) == 1

			/// Using 3rd Byte
			frame.BitRateIndex = maskAndShift(frame.header[2], 0, 4)
			frame.SampleRate = maskAndShift(frame.header[2], 4, 2)
			frame.PaddingBit = maskAndShift(frame.header[2], 6, 1) == 0
			frame.PrivateBit = maskAndShift(frame.header[2], 7, 1) == 0

			// fmt.Printf("\n ** Decoding BitRate %v kbps ", findBitRate(frame.BitRateIndex, frame.AudioVersion, frame.LayerDescription))
			// fmt.Printf("\n ** Decoding SamplingRate  %v Hz", findSampleRate(frame.SampleRate, frame.AudioVersion))

			/// Using 4th Byte
			frame.ChannelMode = maskAndShift(frame.header[3], 0, 2)
			frame.ModeExtension = maskAndShift(frame.header[3], 2, 2)
			frame.CopyRight = maskAndShift(frame.header[3], 4, 1) == 1
			frame.Original = maskAndShift(frame.header[3], 5, 1) == 1
			frame.Emphasis = maskAndShift(frame.header[3], 6, 2)

			frmLength := frame.GetFrameLengthBytes()

			if frmLength == 0 {
				// fmt.Printf("\n ** ZERO SIZE bytes %v ", frmLength)
				frame.FrameSync = false
				/// Invalid
			} else {
				fmt.Printf("\n =============== Frame SYNC found %0X %0X", frame.header[0], frame.header[1])

				fmt.Printf("\n Audio Version %v", frame.AudioVersion)
				fmt.Printf("\n Layer Description %v", frame.LayerDescription)
				fmt.Printf("\n Protection Bit %v", frame.ProtectionBit)
				fmt.Printf("\n frame.BitRateIndex  %v", frame.BitRateIndex)
				fmt.Printf("\n frame.SampleRate %v", frame.SampleRate)
				fmt.Printf("\n frame.PaddingBit %v", frame.PaddingBit)
				fmt.Printf("\n frame.PrivateBit %v", frame.ProtectionBit)

				fmt.Printf("\n frame.ChannelMode %v", frame.ChannelMode)
				fmt.Printf("\n frame.ModeExtension %v", frame.ModeExtension)
				fmt.Printf("\n frame.CopyRight %v", frame.CopyRight)
				fmt.Printf("\n frame.Original  %v", frame.Original)
				fmt.Printf("\n frame.Emphasis %v", frame.Emphasis)
				fmt.Printf("\n Other Info Bit Rate %v bps ", frame.GetBitRate()*1000)
				fmt.Printf("\n Other Info Sampling Frequency %v Hz", frame.GetSampleFreq())

				fmt.Printf("\n ** NON ZERO ? Framelength bytes %v ", frmLength)

			}

		}

	}

	return frame.FrameSync
}

// func main() {
//
// 	bitmask = [8]int{1, 2, 4, 8, 16, 32, 64, 128}
// 	loadBitRateTable()
//
// 	file, err := os.OpenFile("temp.mp3", os.O_RDONLY, 0)
// 	if err != nil {
// 		log.Fatal("Unable to open file temp.mp3", err)
// 		return
// 	}
// 	var frame Mp3Frame
//
// 	var raw []byte = make([]byte, 4)
// 	var bytepos int
// 	for {
// 		n, err := file.Read(raw)
// 		if err != nil {
// 			break
// 		}
//
// 		for j := 0; j < n; j++ {
//
// 			if frame.pushAndValidate(raw[j]) {
// 				// if err != nil {
// 				fmt.Printf("\nFile position : %v ", bytepos+j)
//
// 				time.Sleep(500 * time.Millisecond)
// 				// }
// 			}
// 		}
// 		bytepos += n
// 	}
// 	// var test byte = 183
// 	// fmt.Printf("\n %v \n", maskAndShift(test, 0, 3))
//
// }
