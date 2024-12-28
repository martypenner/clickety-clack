package main

import (
	"clickety-clack/internal/config"
	"clickety-clack/internal/keylistener"
	"clickety-clack/internal/soundplayer"

	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	soundsDir := flag.String("soundsDir", "sounds", "Directory containing soundpacks, each in their own subdirectory")
	flag.Parse()

	// Scan for available sound packs
	soundPacks, err := config.ScanSoundPacks(*soundsDir)
	if err != nil {
		fmt.Printf("Error scanning sound packs: %v\n", err)
		return
	}

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
	keyChan := make(chan string, 100)

	// Initialize with the first sound pack
	currentPlayer, err := soundplayer.NewPlayer(filepath.Join(*soundsDir, soundPacks[0].Directory), soundPacks[0])
	if err != nil {
		fmt.Printf("Error initializing sound player: %v\n", err)
		return
	}

	// Initialize Fyne app
	app := app.New()
	mainWindow := app.NewWindow("Mechvibes")

	// Create volume binding and slider
	volume := binding.NewFloat()
	volume.Set(65) // Default volume
	volumeSlider := widget.NewSliderWithData(0, 100, volume)
	volumeLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(volume, "Volume: %.0f"))

	// Create sound selector with all available sound packs
	var soundPackNames []string
	soundPackMap := make(map[string]*config.SoundPack)
	for _, sp := range soundPacks {
		soundPackNames = append(soundPackNames, sp.Name)
		soundPackMap[sp.Name] = sp
	}

	soundSelector := widget.NewSelect(soundPackNames, func(value string) {
		selectedPack := soundPackMap[value]
		if selectedPack == nil {
			fmt.Printf("Error: Could not find sound pack %s\n", value)
			return
		}

		// Create new player with selected sound pack
		newPlayer, err := soundplayer.NewPlayer(filepath.Join(*soundsDir, selectedPack.Directory), selectedPack)
		if err != nil {
			fmt.Printf("Error switching sound pack: %v\n", err)
			return
		}

		// Clean up the old player before replacing it
		if currentPlayer != nil {
			currentPlayer.Cleanup()
		}

		// Update the current player
		currentPlayer = newPlayer

		// Set the volume on the new player
		if v, err := volume.Get(); err == nil {
			currentPlayer.SetVolume(v)
		}
	})
	soundSelector.SetSelected(soundPacks[0].Name)

	// Create show tray icon checkbox
	showTrayIcon := widget.NewCheck("Show Tray Icon", func(checked bool) {
		fmt.Printf("Show tray icon: %v\n", checked)
	})
	showTrayIcon.SetChecked(true)

	// Create random sound button
	randomSound := widget.NewButton("Set random sound", func() {
		fmt.Println("Random sound selected")
	})

	// Create more sounds button
	moreSounds := widget.NewButton("More sounds…", func() {
		fmt.Println("Opening more sounds…")
	})

	// Layout the UI
	content := container.NewVBox(
		soundSelector,
		container.NewHBox(randomSound, moreSounds),
		volumeLabel,
		volumeSlider,
		showTrayIcon,
	)

	mainWindow.SetContent(content)

	// Set up system tray if supported
	if desk, ok := app.(desktop.App); ok {
		systrayMenu := fyne.NewMenu("Mechvibes",
			fyne.NewMenuItem("Show", func() {
				mainWindow.Show()
			}),
			fyne.NewMenuItem("Quit", func() {
				app.Quit()
			}))
		desk.SetSystemTrayMenu(systrayMenu)
		desk.SetSystemTrayIcon(theme.VolumeUpIcon())
	}

	// Handle volume changes
	go func() {
		for {
			v, err := volume.Get()
			if err == nil {
				currentPlayer.SetVolume(v)
			}
			select {
			case <-ctx.Done():
				return
			default:
				// Continue listening for volume changes
			}
		}
	}()

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
			case keyCode, ok := <-keyChan:
				if !ok {
					fmt.Println("keyChan closed, exiting goroutine...")
					done <- true
					return
				}
				fmt.Printf("Debug: Received key event. Code: %s\n", keyCode)

				err := currentPlayer.PlaySound(keyCode)
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

	mainWindow.Resize(fyne.NewSize(400, 300))
	mainWindow.ShowAndRun()

	<-done
	fmt.Println("Stopping key listener...")
	err = listener.Stop()
	if err != nil {
		fmt.Printf("Error stopping key listener: %v\n", err)
	}
	fmt.Println("\nKeyboard Emulator stopped.")
}
