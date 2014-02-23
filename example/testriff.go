package main

import (
	"fmt"
	"log"
	"os"
	"github.com/wiless/gocodec"
)

func main() {

	filename := "test.wav"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	riffheader := gocodec.ParseFile(filename)
	fmt.Printf("\n%s :  Riff header is  %v", filename, riffheader)

	/// You may modify Header & Write it back 
	riffheader.SampleRate = 8000
	file, err := os.OpenFile(filename, os.O_RDWR, os.ModeSetuid) // For read access.
	if err != nil {
		log.Fatal(err)
	}
	file.WriteAt(riffheader.Bytes(), 0)

}
