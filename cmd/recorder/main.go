package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/jacobidis/gotune/pkg/audio"
)

func main() {
	filename := flag.String("f", "audio.wav", "Filename of recording")
	flag.Parse()
	sampleRate := uint32(48000)
	numChannels := uint32(1)
	writer, err := audio.NewWriter(*filename, sampleRate, numChannels)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer writer.Close()

	ai, err := audio.NewAudioInput(sampleRate)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer ai.Stop()

	if err := ai.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, closing file...")
		ai.Stop()
	}()

	for frame := range ai.Frames {
		if err := writer.WriteFrame(frame); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
