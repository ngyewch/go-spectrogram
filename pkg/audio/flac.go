package audio

import (
	"github.com/mewkiz/flac"
	"io"
	"os"
)

type FLACSource struct {
	info   Info
	frames [][]float64
}

func ReadFLACFromFile(path string) (*FLACSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadFLACFromReader(f)
}

func ReadFLACFromReader(reader io.Reader) (*FLACSource, error) {
	stream, err := flac.Parse(reader)
	if err != nil {
		return nil, err
	}

	flacFile := FLACSource{
		info: Info{
			NumChannels:   int(stream.Info.NChannels),
			SampleRate:    int(stream.Info.SampleRate),
			BitsPerSample: int(stream.Info.BitsPerSample),
		},
		frames: make([][]float64, 0),
	}

	s := 1 << (stream.Info.BitsPerSample - 1)
	q := 1 / float64(s)
	for true {
		src, err := stream.ParseNext()
		if err == io.EOF {
			break
		}
		n := len(src.Subframes[0].Samples)
		dst := make([][]float64, n)
		for i := 0; i < n; i++ {
			dst[i] = make([]float64, int(stream.Info.NChannels))
			for j := 0; j < int(stream.Info.NChannels); j++ {
				dst[i][j] = float64(src.Subframes[j].Samples[i]) * q
			}
		}
		flacFile.frames = append(flacFile.frames, dst...)
	}

	return &flacFile, nil
}

func (f *FLACSource) Info() Info {
	return f.info
}

func (f *FLACSource) Frames() [][]float64 {
	return f.frames
}
