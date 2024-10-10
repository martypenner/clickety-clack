package soundplayer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

type BeepPlayer struct {
	sounds       map[int]*beep.Buffer
	defaultSound *beep.Buffer
	volume       float64
}

func NewPlayer(soundsDir string) (Player, error) {
	player := &BeepPlayer{
		sounds: make(map[int]*beep.Buffer),
		volume: 100, // Classic 0 to 100 percent volume
	}

	err := speaker.Init(44100, 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize speaker: %v", err)
	}

	err = filepath.Walk(soundsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".wav") {
			keyCode := int(strings.ToLower(info.Name())[0])
			sound, err := loadSound(path)
			if err != nil {
				fmt.Printf("Warning: failed to load sound %s: %v\n", path, err)
				return nil
			}
			player.sounds[keyCode] = sound

			if player.defaultSound == nil {
				player.defaultSound = sound
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load sounds from directory: %v", err)
	}

	if player.defaultSound == nil {
		return nil, fmt.Errorf("failed to load any sounds")
	}

	return player, nil
}

func (p *BeepPlayer) PlaySound(keyCode int) error {
	sound, ok := p.sounds[keyCode]
	if !ok {
		sound = p.defaultSound
	}

	if sound == nil {
		return fmt.Errorf("no sound available for key code %d", keyCode)
	}

	// Calculate volume effect based on player's volume (0-200)
	// 0 on player = 0% volume (-2 on effects)
	// 100 on player = 100% volume (0 on effects)
	// 200 on player = 200% volume (2 on effects)
	maxVolume := 2.0
	minVolume := -2.0
	volume := (p.volume/100.0 - 1) * 2 // Linear mapping from 0-200 to -2-2
	volume = math.Max(minVolume, math.Min(maxVolume, volume))
	volumeEffect := &effects.Volume{
		Streamer: sound.Streamer(0, sound.Len()),
		Base:     2,
		Volume:   volume,
		Silent:   volume == minVolume,
	}

	speaker.Play(volumeEffect)
	return nil
}

func (p *BeepPlayer) SetVolume(volume float64) {
	p.volume = volume
}

func loadSound(filepath string) (*beep.Buffer, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	return buffer, nil
}

type Player interface {
	PlaySound(keyCode int) error
	SetVolume(volume float64)
}
