package audio

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"os"
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
	mimeType, err := DetectMimeTypeFromFile(path)
	if err != nil {
		return nil, err
	}

	if mimeType == "audio/wav" {
		return ReadWAVFromFile(path)
	} else if mimeType == "audio/flac" {
		return ReadFLACFromFile(path)
	}

	return nil, fmt.Errorf("unsupported MIME type: %s", mimeType)
}

func DetectMimeTypeFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return DetectMimeTypeFromReader(f)
}

func DetectMimeTypeFromReader(reader io.Reader) (string, error) {
	fileMimeType, err := mimetype.DetectReader(reader)
	if err != nil {
		return "", err
	}

	return fileMimeType.String(), nil
}