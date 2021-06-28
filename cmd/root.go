package cmd

import (
	"fmt"
	"github.com/ngyewch/go-spectrogram/audio"
	"github.com/ngyewch/go-spectrogram/spectrogram"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

var (
	rootCmd = &cobra.Command{
		Use:   "spectrogram [flags] input_audio_path output_image_path",
		Short: "spectrogram.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := run(cmd, args)
			if err != nil {
				panic(fmt.Errorf("Fatal error: %s \n", err))
			}
		},
	}

	channel              uint
	fftSamples           uint
	overlap              uint
	windowFunctionName   string
	minFrequency         uint
	maxFrequency         uint
	relativeMinFrequency float64
	relativeMaxFrequency float64
	relativeMinDecibels  float64
	relativeMaxDecibels  float64
	colorMapName         string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	windowFunction := spectrogram.GetWindowFunctionByName(windowFunctionName)
	if windowFunction == nil {
		return fmt.Errorf("unknown window function: %s", windowFunctionName)
	}

	colorMap := spectrogram.GetColorMapByName(colorMapName)
	if colorMap == nil {
		return fmt.Errorf("unknown color map: %s", colorMapName)
	}

	spectrogramOptions := spectrogram.SpectrogramOptions{
		Channel:        channel,
		FftSamples:     fftSamples,
		Overlap:        &overlap,
		WindowFunction: windowFunction,
	}
	renderOptions := spectrogram.RenderOptions{
		ColorMap: colorMap,
	}
	if isFlagPassed(cmd.Flags(), "min-freq") {
		renderOptions.MinFrequency = &minFrequency
	}
	if isFlagPassed(cmd.Flags(), "max-freq") {
		renderOptions.MaxFrequency = &maxFrequency
	}
	if isFlagPassed(cmd.Flags(), "relative-min-freq") {
		renderOptions.RelativeMinFrequency = &relativeMinFrequency
	}
	if isFlagPassed(cmd.Flags(), "relative-max-freq") {
		renderOptions.RelativeMaxFrequency = &relativeMaxFrequency
	}
	if isFlagPassed(cmd.Flags(), "relative-min-db") {
		renderOptions.RelativeMinDecibels = &relativeMinDecibels
	}
	if isFlagPassed(cmd.Flags(), "relative-max-db") {
		renderOptions.RelativeMaxDecibels = &relativeMaxDecibels
	}

	inputPath := args[0]
	outputPath := args[1]

	src, err := audio.ReadFromFile(inputPath)
	if err != nil {
		return err
	}

	spec, err := spectrogram.GenerateSpectrogram(src, spectrogramOptions)
	if err != nil {
		return err
	}

	img, _, err := spec.ToImage(renderOptions)
	if err != nil {
		return err
	}

	err = saveImageToFile(img, outputPath)

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().UintVar(&channel, "channel", 0, "Channel.")
	rootCmd.Flags().UintVar(&fftSamples, "fft-samples", 1024, "FFT samples.")
	rootCmd.Flags().UintVar(&overlap, "overlap", 768, "Overlap.")
	rootCmd.Flags().StringVar(&windowFunctionName, "window-func", "hann", "Window function.")
	rootCmd.Flags().UintVar(&minFrequency, "min-freq", 0, "Min frequency.")
	rootCmd.Flags().UintVar(&maxFrequency, "max-freq", 0, "Max frequency.")
	rootCmd.Flags().Float64Var(&relativeMinFrequency, "relative-min-freq", 0, "Relative min frequency.")
	rootCmd.Flags().Float64Var(&relativeMaxFrequency, "relative-max-freq", 0, "Relative max frequency.")
	rootCmd.Flags().Float64Var(&relativeMinDecibels, "relative-min-db", 0, "Relative min decibels.")
	rootCmd.Flags().Float64Var(&relativeMaxDecibels, "relative-max-db", 0, "Relative max decibels.")
	rootCmd.Flags().StringVar(&colorMapName, "color-map", "inferno", "Color map.")
}

func initConfig() {
	// do nothing
}

func isFlagPassed(flagSet *pflag.FlagSet, name string) bool {
	found := false
	flagSet.Visit(func(f *pflag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func saveImageToFile(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := filepath.Ext(path)
	if ext == ".png" {
		err = png.Encode(f, img)
		if err != nil {
			return err
		}
	} else if ext == ".jpg" || ext == ".jpeg" {
		err = jpeg.Encode(f, img, nil)
		if err != nil {
			return err
		}
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
