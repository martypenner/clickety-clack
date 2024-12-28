package config

import (
	"encoding/json"
	"os"
)

type SoundPack struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ConfigType     string `json:"sound_key_type" validate:"oneof=single multiple"`
	IncludesNumpad bool   `json:"includes_numpad"`
	SoundDemo      string `json:"sound"`
	// This handles multiple ranges of key codes so we can operate across
	// different OSes:
	// Standard PS/2 scancodes (lower numbers like 19)
	// Linux evdev codes (3000+ range)
	// macOS NSEvent keyCodes (61000+ range)
	Defines map[string]*string `json:"defines,omitempty"`
}

func LoadSoundPack(filename string) (*SoundPack, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config SoundPack
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (sp *SoundPack) GetSoundFileName(keyCode string) *string {
	if soundFile, exists := sp.Defines[keyCode]; exists && soundFile != nil {
		return soundFile
	}

	// If the sound file is null or doesn't exist, find the first non-null sound
	for _, sound := range sp.Defines {
		if sound != nil {
			return sound
		}
	}
	return nil
}
