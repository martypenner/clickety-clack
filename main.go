package main

import (
	"clickety-clack/internal/keylistener"
	"clickety-clack/internal/soundplayer"

	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
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

	// Initialize Fyne app
	app := app.New()
	mainWindow := app.NewWindow("Mechvibes")

	// Create volume binding and slider
	volume := binding.NewFloat()
	volume.Set(65) // Default volume
	volumeSlider := widget.NewSliderWithData(0, 100, volume)
	volumeLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(volume, "Volume: %.0f"))

	// Create sound selector (placeholder for now)
	soundSelector := widget.NewSelect([]string{"YUNZII C68 - AnDr3W"}, func(value string) {
		fmt.Printf("Selected sound pack: %s\n", value)
	})
	soundSelector.SetSelected("YUNZII C68 - AnDr3W")

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
	moreSounds := widget.NewButton("More sounds...", func() {
		fmt.Println("Opening more sounds...")
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
				player.SetVolume(v)
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
