package keylistener

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	hook "github.com/robotn/gohook"
)

type KeyListener interface {
	Start(ctx context.Context, ch chan<- string) error
	Stop() error
}

type gohookKeyListener struct {
	mutex       sync.Mutex
	started     bool
	eventsChan  chan hook.Event
	cancelFunc  context.CancelFunc
	pressedKeys map[uint16]bool
}

func NewGohookKeyListener() KeyListener {
	return &gohookKeyListener{
		pressedKeys: make(map[uint16]bool),
	}
}

func (g *gohookKeyListener) Start(ctx context.Context, ch chan<- string) error {
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
			g.mutex.Lock()
			switch event.Kind {
			case hook.KeyDown:
				if !g.pressedKeys[event.Rawcode] {
					g.pressedKeys[event.Rawcode] = true
					keyCode := strconv.Itoa(int(event.Rawcode))
					fmt.Printf("Debug: Sending key code: %s\n", keyCode)
					ch <- keyCode
				}
			case hook.KeyUp:
				delete(g.pressedKeys, event.Rawcode)
			}
			g.mutex.Unlock()
		}
	}()

	return nil
}

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
