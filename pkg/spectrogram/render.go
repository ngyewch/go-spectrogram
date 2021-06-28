package spectrogram

import (
	"errors"
	"github.com/montanaflynn/stats"
	"image"
	"image/color"
	"math"
)

type RenderOptions struct {
	MinFrequency         *uint
	MaxFrequency         *uint
	RelativeMinFrequency *float64
	RelativeMaxFrequency *float64
	RelativeMinDecibels  *float64
	RelativeMaxDecibels  *float64
	ColorMap             []color.Color
}

type RenderInfo struct {
	MinFrequency uint
	MaxFrequency uint
	MinDb        float64
	MaxDb        float64
}

func (spectrogram *Spectrogram) ToImage(options RenderOptions) (image.Image, *RenderInfo, error) {
	sampleRate := spectrogram.SampleRate
	fsOver2 := float64(sampleRate) / 2
	if (options.MinFrequency != nil) && (options.RelativeMinFrequency != nil) {
		return nil, nil, errors.New("cannot specify both MinFrequency and RelativeMinFrequency")
	}
	if options.MinFrequency != nil {
		if (*options.MinFrequency < 0) || (*options.MinFrequency > uint(fsOver2)) {
			return nil, nil, errors.New("invalid MinFrequency")
		}
	}
	if options.RelativeMinFrequency != nil {
		if (*options.RelativeMinFrequency < 0) || (*options.RelativeMinFrequency > 1) {
			return nil, nil, errors.New("invalid RelativeMinFrequency")
		}
	}

	if (options.MaxFrequency != nil) && (options.RelativeMaxFrequency != nil) {
		return nil, nil, errors.New("cannot specify both MaxFrequency and RelativeMaxFrequency")
	}
	if options.MaxFrequency != nil {
		if (*options.MaxFrequency < 0) || (*options.MaxFrequency > uint(fsOver2)) {
			return nil, nil, errors.New("invalid MaxFrequency")
		}
	}
	if options.RelativeMaxFrequency != nil {
		if (*options.RelativeMaxFrequency < 0) || (*options.RelativeMaxFrequency > 1) {
			return nil, nil, errors.New("invalid RelativeMaxFrequency")
		}
	}

	var minFreq float64 = 0
	if options.MinFrequency != nil {
		minFreq = float64(*options.MinFrequency)
	} else if options.RelativeMinFrequency != nil {
		minFreq = *options.RelativeMinFrequency * fsOver2
	}

	maxFreq := float64(spectrogram.SampleRate / 2)
	if options.MaxFrequency != nil {
		maxFreq = float64(*options.MaxFrequency)
	} else if options.RelativeMaxFrequency != nil {
		maxFreq = *options.RelativeMaxFrequency * fsOver2
	}

	if minFreq >= maxFreq {
		return nil, nil, errors.New("minFrequency must be less than maxFrequency")
	}

	minIndexRatio := minFreq / fsOver2
	maxIndexRatio := maxFreq / fsOver2
	minIndex := int(math.Floor(minIndexRatio * float64(spectrogram.FftSamples/2)))
	maxIndex := int(math.Min(math.Ceil(maxIndexRatio*float64(spectrogram.FftSamples/2)), float64((spectrogram.FftSamples/2)-1)))

	statsValues := make([]float64, 0)
	for i := 0; i < len(spectrogram.Data); i++ {
		specColumn := spectrogram.Data[i]
		statsValues = append(statsValues, specColumn[minIndex:maxIndex+1]...)
	}

	var median float64
	if (options.RelativeMinDecibels != nil) || (options.RelativeMaxDecibels != nil) {
		_median, err := stats.Median(statsValues)
		if err != nil {
			return nil, nil, err
		}
		median = _median
	}

	var minDb float64
	if options.RelativeMinDecibels != nil {
		minDb = median + *options.RelativeMinDecibels
	} else {
		_minDb, err := stats.Min(statsValues)
		if err != nil {
			return nil, nil, err
		}
		minDb = _minDb
	}

	var maxDb float64
	if options.RelativeMaxDecibels != nil {
		maxDb = median + *options.RelativeMaxDecibels
	} else {
		_maxDb, err := stats.Max(statsValues)
		if err != nil {
			return nil, nil, err
		}
		maxDb = _maxDb
	}

	dbRange := maxDb - minDb

	imageHeight := maxIndex - minIndex + 1
	img := image.NewNRGBA(image.Rect(0, 0, len(spectrogram.Data), imageHeight))
	for x := 0; x < len(spectrogram.Data); x++ {
		specColumn := spectrogram.Data[x]
		for y := minIndex; y <= maxIndex; y++ {
			spec := math.Min(math.Max(specColumn[y], minDb), maxDb)
			normalizedSpec := (spec - minDb) / dbRange
			colorIndex := int(math.Round(normalizedSpec * float64(len(options.ColorMap)-1)))
			if colorIndex < 0 {
				colorIndex = 0
			}
			if colorIndex >= len(options.ColorMap) {
				colorIndex = len(options.ColorMap) - 1
			}
			c := options.ColorMap[colorIndex]
			img.Set(x, imageHeight-(y-minIndex)-1, c)
		}
	}

	return img, &RenderInfo{
		MinFrequency: uint(minFreq),
		MaxFrequency: uint(maxFreq),
		MinDb:        minDb,
		MaxDb:        maxDb,
	}, nil
}
