package soundplayer

import (
	"clickety-clack/internal/config"

	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

var speakerInitialized bool

type BeepPlayer struct {
	sounds       map[string]*beep.Buffer
	defaultSound *beep.Buffer
	volume       float64
	mutex        sync.Mutex
}

func NewPlayer(soundsDir string, soundPack *config.SoundPack) (Player, error) {
	player := &BeepPlayer{
		sounds: make(map[string]*beep.Buffer),
		volume: 100, // Classic 0 to 100 percent volume
	}

	if !speakerInitialized {
		err := speaker.Init(44100, 1024)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize speaker: %v", err)
		}
		speakerInitialized = true
	}

	// Load only the sounds defined in the config
	for keyCodeStr, soundFile := range soundPack.Defines {
		if soundFile == nil {
			continue
		}

		soundPath := filepath.Join(soundsDir, *soundFile)
		sound, err := loadSound(soundPath)
		if err != nil {
			fmt.Printf("Warning: failed to load sound %s: %v\n", soundPath, err)
			continue
		}

		player.sounds[keyCodeStr] = sound

		if player.defaultSound == nil {
			player.defaultSound = sound
		}
	}

	if player.defaultSound == nil {
		return nil, fmt.Errorf("failed to load any sounds")
	}

	return player, nil
}

func (p *BeepPlayer) PlaySound(keyCode string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.sounds == nil {
		return fmt.Errorf("player has been cleaned up")
	}

	sound, ok := p.sounds[keyCode]
	if !ok {
		sound = p.defaultSound
	}

	if sound == nil {
		return fmt.Errorf("no sound available for key code %s", keyCode)
	}

	// Offset the sound by 20ms to account for the delay in many sound files.
	// Without this, we're not lined up with mechvibes.
	// Convert ms to samples (assuming 44100Hz sample rate).
	// 15ms seemed to be the sweet spot.
	offsetMs := 15 // 15 milliseconds offset
	samplesPerMs := 44100 / 1000
	offsetSamples := offsetMs * samplesPerMs

	// Calculate volume effect based on player's volume (0-200)
	// 0 on player = 0% volume (-2 on effects)
	// 100 on player = 100% volume (0 on effects)
	// 200 on player = 200% volume (2 on effects)
	maxVolume := 2.0
	minVolume := -2.0
	volume := (p.volume/100.0 - 1) * 2 // Linear mapping from 0-200 to -2-2
	volume = math.Max(minVolume, math.Min(maxVolume, volume))
	volumeEffect := &effects.Volume{
		Streamer: sound.Streamer(offsetSamples, sound.Len()),
		Base:     2,
		Volume:   volume,
		Silent:   volume == minVolume,
	}

	speaker.Play(volumeEffect)
	return nil
}

func (p *BeepPlayer) SetVolume(volume float64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.volume = volume
}

func (p *BeepPlayer) Cleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.sounds = nil
	p.defaultSound = nil

	// Let the garbage collector handle the actual cleanup of the buffers
}

func loadSound(filepath string) (*beep.Buffer, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var streamer beep.StreamSeekCloser
	var format beep.Format

	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])
	switch ext {
	case "ogg":
		streamer, format, err = vorbis.Decode(f)
	case "wav":
		streamer, format, err = wav.Decode(f)
	case "mp3":
		streamer, format, err = mp3.Decode(f)
	case "flac":
		streamer, format, err = flac.Decode(f)
	default:
		return nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %v", ext, err)
	}
	defer streamer.Close()

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	return buffer, nil
}

type Player interface {
	PlaySound(keyCode string) error
	SetVolume(volume float64)
	Cleanup()
}
