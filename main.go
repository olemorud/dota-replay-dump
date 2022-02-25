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

	size, err := parse.First(r)
	if err != nil {
		fmt.Println(err)
	}

	if _, err := r.Discard(int(size) - 12); err != nil {
		fmt.Println(fmt.Errorf("error jumping to last frame: %v", err))
	}

	frame, err := parse.DecodeNextFrame(r)
	if err != nil {
		fmt.Println(fmt.Errorf("error parsing frame: %v", err))
	}

	//fmt.Printf("%+v", frame.Message)

	for _, player := range frame.Message.GameInfo.Dota.PlayerInfo {
		fmt.Println(*player.HeroName)
		fmt.Println(*player.PlayerName)
		fmt.Println("Steam:", *player.Steamid)
		fmt.Println("")

	}

	//fmt.Printf("%+v", )

}
