package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/olemorud/replay-parser/parse"
)

func main() {
	f, err := os.Open("replays/dog2.dem")
	if err != nil {
		fmt.Println(fmt.Errorf("error reading file: %v", err))
	}

	r := bufio.NewReader(f)
	r.Discard(16)

	for i, err := r.Peek(1); len(i) == 1 && err == nil; {

		frame, err := parse.DecodeNextFrame(r)
		if err != nil {
			fmt.Println(fmt.Errorf("error parsing frame: %v", err))
			return
		}
		if frame != nil {
			fmt.Printf("message: %+v", frame.Message)
			fmt.Printf("\n\n")
		}
	}

	// offset, err := parse.First(r)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// _ = offset // remove later

	// if _, err := r.Discard(int(offset) - 12); err != nil {
	// 	fmt.Println(fmt.Errorf("error jumping to last frame: %v", err))
	// }
}
