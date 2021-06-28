package audio

import (
	"github.com/ngyewch/go-spectrogram/wave"
	"io"
	"os"
)

type WAVSource struct {
	wav *wave.Wave
}

func ReadWAVFromFile(path string) (*WAVSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadWAVFromReader(f)
}

func ReadWAVFromReader(reader io.Reader) (*WAVSource, error) {
	wav, err := wave.ReadWaveFromReader(reader, false)
	if err != nil {
		return nil, err
	}
	return &WAVSource{
		wav: wav,
	}, nil
}

func (wavSource *WAVSource) Info() Info {
	return Info{
		NumChannels:   wavSource.wav.Fmt.NumChannels,
		SampleRate:    wavSource.wav.Fmt.SampleRate,
		BitsPerSample: wavSource.wav.Fmt.BitsPerSample,
	}
}

func (wavSource *WAVSource) Frames() [][]float64 {
	return wavSource.wav.Frames
}
