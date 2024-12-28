package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	// Directory path relative to sounds directory
	Directory string `json:"-"`
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

	// Set the directory to be relative to the sounds directory
	config.Directory = filepath.Base(filepath.Dir(filename))

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

func ScanSoundPacks(soundsDir string) ([]*SoundPack, error) {
	var soundPacks []*SoundPack

	err := filepath.Walk(soundsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Base(path) == "config.json" {
			soundPack, err := LoadSoundPack(path)
			if err != nil {
				fmt.Printf("Warning: Failed to load sound pack at %s: %v\n", path, err)
				return nil
			}

			soundPacks = append(soundPacks, soundPack)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan sound packs: %v", err)
	}

	if len(soundPacks) == 0 {
		return nil, fmt.Errorf("no valid sound packs found in %s", soundsDir)
	}

	return soundPacks, nil
}
