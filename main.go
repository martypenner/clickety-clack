package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"clickety-clack/keylistener"
	"clickety-clack/soundplayer"
)

func main() {
	soundsDir := flag.String("sounds", "sounds", "Directory containing sound files")
	flag.Parse()

	// Create a new KeyListener using gohook
	listener := keylistener.NewGohookKeyListener()
	if listener == nil {
		fmt.Println("Failed to initialize global key listener.")
		return
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a buffered channel with a capacity of 100
	keyChan := make(chan rune, 100)

	player, err := soundplayer.NewPlayer(*soundsDir)
	if err != nil {
		fmt.Printf("Error initializing sound player: %v\n", err)
		return
	}

	fmt.Println("Starting key listener...")
	err = listener.Start(ctx, keyChan)
	if err != nil {
		fmt.Printf("Error starting key listener: %v\n", err)
		return
	}
	fmt.Println("Key listener started successfully.")

	// Handle graceful termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Keyboard Emulator started. Press keys to see them printed. Press Ctrl+C to exit.")

	done := make(chan bool)

	go func() {
		fmt.Println("Goroutine started, waiting for key events...")
		for {
			select {
			case key, ok := <-keyChan:
				if !ok {
					fmt.Println("keyChan closed, exiting goroutine...")
					done <- true
					return
				}
				fmt.Printf("Debug: Received key event. Key: %q (Code: %d)\n", key, key)

				err := player.PlaySound(int(key))
				if err != nil {
					fmt.Printf("Error playing sound: %v\n", err)
				}

			case sig := <-sigs:
				fmt.Printf("\nReceived signal: %v. Shutting down...\n", sig)
				cancel()
				done <- true
				return
			case <-ctx.Done():
				fmt.Println("Context cancelled, exiting goroutine...")
				done <- true
				return
			}
		}
	}()

	<-done
	fmt.Println("Stopping key listener...")
	err = listener.Stop()
	if err != nil {
		fmt.Printf("Error stopping key listener: %v\n", err)
	}
	fmt.Println("\nKeyboard Emulator stopped.")
}
