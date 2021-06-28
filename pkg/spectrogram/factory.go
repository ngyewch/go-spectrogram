package spectrogram

import (
	"github.com/dim13/colormap"
	"github.com/mjibson/go-dsp/window"
	"image/color"
)

func GetWindowFunctionByName(name string) func(int) []float64 {
	if name == "hann" {
		return window.Hann
	} else if name == "hamming" {
		return window.Hamming
	} else if name == "bartlett" {
		return window.Bartlett
	} else if name == "blackman" {
		return window.Blackman
	} else if name == "flatTop" {
		return window.FlatTop
	} else if name == "rectangular" {
		return window.Rectangular
	} else {
		return nil
	}
}

func GetColorMapByName(name string) []color.Color {
	if name == "inferno" {
		return colormap.Inferno
	} else if name == "magma" {
		return colormap.Magma
	} else if name == "plasma" {
		return colormap.Plasma
	} else if name == "viridis" {
		return colormap.Viridis
	} else {
		return nil
	}
}
