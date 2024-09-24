package soundplayer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

type BeepPlayer struct {
	sounds       map[int]*beep.Buffer
	defaultSound *beep.Buffer
}

func NewPlayer(soundsDir string) (Player, error) {
	player := &BeepPlayer{
		sounds: make(map[int]*beep.Buffer),
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

	speaker.Play(beep.Seq(sound.Streamer(0, sound.Len())))
	return nil
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
}
