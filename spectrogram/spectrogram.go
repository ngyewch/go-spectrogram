package spectrogram

import (
	"errors"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
	"github.com/ngyewch/go-spectrogram/audio"
	"math"
)

type WindowFunction func(int) []float64

type SpectrogramOptions struct {
	Channel        uint
	FftSamples     uint
	Overlap        *uint
	Segments       *uint
	WindowFunction WindowFunction
}

type Spectrogram struct {
	SampleRate  uint
	NumChannels uint
	FftSamples  uint
	Data        [][]float64
}

func GenerateSpectrogram(audioFile audio.Source, options SpectrogramOptions) (*Spectrogram, error) {
	info := audioFile.Info()
	frames := audioFile.Frames()

	channel := int(options.Channel)
	if channel < 0 || channel >= info.NumChannels {
		return nil, errors.New("invalid channel number")
	}

	fftSamples := options.FftSamples
	if !IsPowerOfTwo(fftSamples) {
		return nil, errors.New("fftSamples must be a power of 2")
	}

	overlap := 0
	hop := 0
	if (options.Segments != nil) && (options.Overlap != nil) {
		return nil, errors.New("cannot specify both Segments and Overlap")
	} else if (options.Segments == nil) && (options.Overlap == nil) {
		overlap = 0
		hop = int(fftSamples)
	} else if (options.Segments == nil) && (options.Overlap != nil) {
		if *options.Overlap >= fftSamples {
			return nil, errors.New("overlap must be less than fftSamples")
		}
		overlap = int(*options.Overlap)
		hop = int(fftSamples) - overlap
	} else if (options.Segments != nil) && (options.Overlap == nil) {
		if *options.Segments <= 1 {
			return nil, errors.New("segments must be greater than 1")
		}
		hop = (len(frames) - int(fftSamples)) / int(*options.Segments-1)
		if hop < 1 {
			hop = 1
		} else if hop > int(fftSamples) {
			hop = int(fftSamples)
		}
		overlap = int(fftSamples) - hop
	}

	bSi := 2 / float64(fftSamples)
	buffer := make([]float64, int(fftSamples))
	specColumns := make([][]float64, 0)
	for i := 0; (i + int(fftSamples)) <= len(frames); i += hop {
		start := i
		end := start + int(fftSamples)
		currentFrame := frames[start:end]
		for j := 0; j < int(fftSamples); j++ {
			buffer[j] = currentFrame[j][channel]
		}
		window.Apply(buffer, options.WindowFunction)
		fftResult := fft.FFTReal(buffer)
		specColumn := make([]float64, len(buffer)/2)
		for j := 0; j < len(buffer)/2; j++ {
			val := fftResult[j]
			mag := math.Sqrt((real(val)*real(val))+(imag(val)*imag(val))) * bSi
			specColumn[j] = 20 * math.Log10(mag)
		}
		specColumns = append(specColumns, specColumn)
	}

	return &Spectrogram{
		SampleRate:  uint(info.SampleRate),
		NumChannels: uint(info.NumChannels),
		FftSamples:  fftSamples,
		Data:        specColumns,
	}, nil
}
