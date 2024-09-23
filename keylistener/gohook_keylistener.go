package keylistener

import (
	"context"
	"fmt"
	"sync"

	hook "github.com/robotn/gohook"
)

type KeyListener interface {
	Start(ctx context.Context, ch chan<- rune) error
	Stop() error
}

type gohookKeyListener struct {
	mutex      sync.Mutex
	started    bool
	eventsChan chan hook.Event
	cancelFunc context.CancelFunc
}

// NewGohookKeyListener creates a new KeyListener using gohook
func NewGohookKeyListener() KeyListener {
	return &gohookKeyListener{}
}

// Start begins listening for global key presses
func (g *gohookKeyListener) Start(ctx context.Context, ch chan<- rune) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.started {
		return fmt.Errorf("key listener already started")
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	g.cancelFunc = cancel

	// Start the hook in a separate goroutine
	g.eventsChan = hook.Start()
	g.started = true

	// Listen for context cancellation to stop the hook
	go func() {
		<-ctx.Done()
		g.Stop()
	}()

	go func() {
		for event := range g.eventsChan {
			if event.Kind != hook.KeyDown {
				continue
			}

			fmt.Printf("Debug: Received key event: %+v\n", event)
			keyRune, ok := keyEventToRune(event)
			if ok {
				fmt.Printf("Debug: Sending key rune: %q\n", keyRune)
				ch <- keyRune
			} else {
				fmt.Println("Debug: Failed to convert key event to rune")
			}
		}
	}()

	return nil
}

// Stop stops listening for global key presses
func (g *gohookKeyListener) Stop() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.started {
		return fmt.Errorf("key listener not started")
	}

	// Stop the hook
	hook.End()
	if g.cancelFunc != nil {
		g.cancelFunc()
	}
	g.started = false
	return nil
}

// keyEventToRune converts a gohook.Event to a rune, if possible
func keyEventToRune(e hook.Event) (rune, bool) {
	// Attempt to convert key to rune
	if e.Keychar != 0 {
		r := rune(e.Keychar)
		return r, true
	}

	// Handle special keys
	switch e.Rawcode {
	case 36: // Return
		return '\n', true
	case 51: // Delete
		return '\b', true
	case 56: // Left Alt
		return '\x1A', true // Substitute character for Left Alt
	case 57: // Caps Lock
		return '\x03', true // End of Text for Caps Lock
	case 58: // Left Ctrl
		return '\x02', true // Start of Text for Left Ctrl
	case 59: // Left Shift
		return '\x10', true // Data Link Escape for Left Shift
	case 60: // Left Win
		return '\x1B', true // Escape for Left Win
	case 61: // Right Shift
		return '\x11', true // Device Control 1 for Right Shift
	case 62: // Right Ctrl
		return '\x04', true // End of Transmission for Right Ctrl
	case 63: // Right Alt
		return '\x1B', true // Escape for Right Alt
	case 64: // Right Win
		return '\x1B', true // Escape for Right Win
	case 65: // Space
		return ' ', true
	case 66: // Pause
		return '\x1B', true // Escape for Pause
	case 67: // Home
		return '\x1B', true // Escape for Home
	case 68: // Up
		return '\x1B', true // Escape for Up Arrow
	case 69: // Page Up
		return '\x1B', true // Escape for Page Up
		// Add more special key mappings as needed
	}

	fmt.Printf("Debug: Unhandled key event: %+v\n", e) // Add this line for debugging
	return 0, false
}
