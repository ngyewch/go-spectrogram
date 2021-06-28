package audio

import (
	"fmt"
	"path/filepath"
)

type Source interface {
	Info() Info
	Frames() [][]float64
}

type Info struct {
	NumChannels   int
	SampleRate    int
	BitsPerSample int
}

func ReadFromFile(path string) (Source, error) {
	ext := filepath.Ext(path)

	if ext == ".wav" {
		return ReadWAVFromFile(path)
	} else if ext == ".flac" {
		return ReadFLACFromFile(path)
	}

	return nil, fmt.Errorf("unsupported file extension: %s", ext)
}
