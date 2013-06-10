package main

import (
	"../geyefi"

	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	fmt.Println("Hello, world!")

	tempDir, err := ioutil.TempDir("", "eyefi")
	if err != nil {
		log.Fatal(err)
	}

	handler := &geyefi.SaveFileHandler{Directory: tempDir}
	log.Printf("Files will be saved to: %s\n", tempDir)
	e := geyefi.NewServer("818b6183a1a0839d88366f5d7a4b0161", handler)
	e.ListenAndServe()
}
