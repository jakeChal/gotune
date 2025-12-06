package main

import (
	"fmt"

	"gitlab.com/jacobidis/gotune/pkg/audio"
	"gitlab.com/jacobidis/gotune/pkg/dsp"
)

func main() {
	const sampleRate = 48000

	ai, err := audio.NewAudioInput(sampleRate)
	if err != nil {
		panic(err)
	}

	if err := ai.Start(); err != nil {
		panic(err)
	}
	defer ai.Stop()

	fmt.Println("Listening...")

	for frame := range ai.Frames {
		val := dsp.PassThrough(frame) // later replace with pitch estimate
		fmt.Printf("\rEnergy: %.2f      ", val)
	}
}
