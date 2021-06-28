package wave

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	FormatPCM        = 0x0001
	FormatIEEEFloat  = 0x0003
	FormatALaw       = 0x0006
	FormatMuLaw      = 0x0007
	FormatExtensible = 0xfffe
)

var (
	riffChunkId    = []byte("RIFF")
	rifxChunkId    = []byte("RIFX")
	waveFormatId   = []byte("WAVE")
	fmtSubChunkId  = []byte("fmt ")
	factSubChunkId = []byte("fact")
	dataSubChunkId = []byte("data")
)

type waveReader struct {
	reader    io.Reader
	byteOrder binary.ByteOrder
}

type sampleReader struct {
	waveFmt        *WaveFmt
	byteOrder      binary.ByteOrder
	bytesPerSample int
	divisor        float64
	midpoint       float64
}

func ReadWaveFromFile(f string, headerOnly bool) (*Wave, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ReadWaveFromReader(file, headerOnly)
}

func ReadWaveFromReader(reader io.Reader, headerOnly bool) (*Wave, error) {
	chunkId, err := readBytes(reader, 4)
	if err != nil {
		return nil, err
	}

	var byteOrder binary.ByteOrder
	if bytes.Compare(chunkId, riffChunkId) == 0 {
		byteOrder = binary.LittleEndian
	} else if bytes.Compare(chunkId, rifxChunkId) == 0 {
		byteOrder = binary.BigEndian
	} else {
		return nil, errors.New(fmt.Sprintf("unknown chunk ID: %s", string(chunkId)))
	}

	var chunkSize uint32
	err = binary.Read(reader, byteOrder, &chunkSize)

	if chunkSize > math.MaxInt32 {
		return nil, errors.New(fmt.Sprintf("%s chunk size too big", string(chunkId)))
	}

	riffReader := io.LimitReader(reader, int64(chunkSize))

	formatId, err := readBytes(riffReader, 4)
	if err != nil {
		return nil, err
	}

	if bytes.Compare(formatId, waveFormatId) != 0 {
		return nil, errors.New(fmt.Sprintf("unknown format ID: %s", string(formatId)))
	}

	waveReader := &waveReader{
		reader:    riffReader,
		byteOrder: byteOrder,
	}

	// handle "fmt " sub-chunk
	subChunkId, subChunkSize, err := waveReader.readSubChunkIdAndSize()
	if err != nil {
		return nil, err
	}
	if bytes.Compare(subChunkId, fmtSubChunkId) != 0 {
		return nil, errors.New(fmt.Sprintf("expected sub-chunk %s, found sub-chunk %s",
			string(fmtSubChunkId), string(subChunkId)))
	}
	waveFmt, err := waveReader.readFmtSubChunk(subChunkSize)
	if err != nil {
		return nil, err
	}

	if headerOnly {
		return &Wave{
			Fmt:    waveFmt,
			Frames: nil,
		}, nil
	}

	// validate supported formats
	switch waveFmt.AudioFormat {
	case FormatPCM:
		break
	case FormatIEEEFloat:
	case FormatALaw:
	case FormatMuLaw:
	case FormatExtensible:
	default:
		return nil, errors.New(fmt.Sprintf("unsupported audio format: 0x%04x", waveFmt.AudioFormat))
	}

	if (waveFmt.BitsPerSample < 8) || (waveFmt.BitsPerSample > 64) || ((waveFmt.BitsPerSample % 8) != 0) {
		return nil, errors.New(fmt.Sprintf("unsupported bits per sample: %d", waveFmt.BitsPerSample))
	}

	// read next sub-chunk
	subChunkId, subChunkSize, err = waveReader.readSubChunkIdAndSize()
	if err != nil {
		return nil, err
	}
	var waveFact *WaveFact = nil
	if bytes.Compare(subChunkId, factSubChunkId) == 0 {
		// handle "fact" sub-chunk
		waveFact, err = waveReader.readFactSubChunk(subChunkSize)
		if err != nil {
			return nil, err
		}

		// read next sub-chunk
		subChunkId, subChunkSize, err = waveReader.readSubChunkIdAndSize()
		if err != nil {
			return nil, err
		}
	}

	// expecting "data" sub-chunk
	if bytes.Compare(subChunkId, dataSubChunkId) != 0 {
		return nil, errors.New(fmt.Sprintf("expected sub-chunk %s, found sub-chunk %s",
			string(dataSubChunkId), string(subChunkId)))
	}

	frames, err := waveReader.readPCMDataSubChunk(subChunkSize, waveFmt)
	if err != nil {
		return nil, err
	}

	return &Wave{
		Fmt:    waveFmt,
		Fact:   waveFact,
		Frames: frames,
	}, nil
}

func (wr *waveReader) readSubChunkIdAndSize() ([]byte, uint32, error) {
	subChunkId, err := readBytes(wr.reader, 4)
	if err != nil {
		return nil, 0, err
	}

	var subChunkSize uint32
	err = binary.Read(wr.reader, wr.byteOrder, &subChunkSize)
	if err != nil {
		return nil, 0, err
	}
	if subChunkSize > math.MaxInt32 {
		return nil, 0, errors.New(fmt.Sprintf("%s sub-chunk size too big", string(subChunkId)))
	}

	return subChunkId, subChunkSize, nil
}

func (wr *waveReader) readFmtSubChunk(subChunkSize uint32) (*WaveFmt, error) {
	reader := io.LimitReader(wr.reader, int64(subChunkSize))

	var waveFmt WaveFmt

	var audioFormat uint16
	err := binary.Read(reader, wr.byteOrder, &audioFormat)
	if err != nil {
		return nil, err
	}
	waveFmt.AudioFormat = int(audioFormat)

	var numChannels uint16
	err = binary.Read(reader, wr.byteOrder, &numChannels)
	if err != nil {
		return nil, err
	}
	waveFmt.NumChannels = int(numChannels)

	var sampleRate uint32
	err = binary.Read(reader, wr.byteOrder, &sampleRate)
	if err != nil {
		return nil, err
	}
	waveFmt.SampleRate = int(sampleRate)

	var byteRate uint32
	err = binary.Read(reader, wr.byteOrder, &byteRate)
	if err != nil {
		return nil, err
	}
	waveFmt.ByteRate = int(byteRate)

	var blockAlign uint16
	err = binary.Read(reader, wr.byteOrder, &blockAlign)
	if err != nil {
		return nil, err
	}
	waveFmt.BlockAlign = int(blockAlign)

	var bitsPerSample uint16
	err = binary.Read(reader, wr.byteOrder, &bitsPerSample)
	if err != nil {
		return nil, err
	}
	waveFmt.BitsPerSample = int(bitsPerSample)

	var extraParamSize uint16
	err = binary.Read(reader, wr.byteOrder, &extraParamSize)
	if err != io.EOF {
		if err != nil {
			return nil, err
		}
		waveFmt.ExtraParamSize = int(extraParamSize)

		if extraParamSize > 0 {
			extraParams, err := readBytes(reader, int(extraParamSize))
			if err != nil {
				return nil, err
			}
			waveFmt.ExtraParams = extraParams
		} else {
			waveFmt.ExtraParams = make([]byte, 0)
		}
	}

	return &waveFmt, nil
}

func (wr *waveReader) readFactSubChunk(subChunkSize uint32) (*WaveFact, error) {
	reader := io.LimitReader(wr.reader, int64(subChunkSize))

	var waveFact WaveFact

	var sampleLength uint32
	err := binary.Read(reader, wr.byteOrder, &sampleLength)
	if err != nil {
		return nil, err
	}
	waveFact.SampleLength = int(sampleLength)

	return &waveFact, nil
}

func (wr *waveReader) readPCMDataSubChunk(subChunkSize uint32, waveFmt *WaveFmt) ([][]float64, error) {
	reader := io.LimitReader(wr.reader, int64(subChunkSize))

	sr := newSampleReader(waveFmt, wr.byteOrder)
	bytesPerSample := waveFmt.BitsPerSample / 8
	buffer := make([]byte, waveFmt.BlockAlign)
	frames := make([][]float64, 0)
	for true {
		_, err := io.ReadFull(reader, buffer)
		if err == io.EOF {
			break
		}
		current := make([]float64, waveFmt.NumChannels)
		for i := 0; i < waveFmt.NumChannels; i++ {
			start := i * bytesPerSample
			end := start + bytesPerSample
			sample := buffer[start:end]
			val, err := sr.readSample(sample)
			if err != nil {
				return nil, err
			}
			current[i] = val
		}
		frames = append(frames, current)
	}

	return frames, nil
}

func newSampleReader(waveFmt *WaveFmt, byteOrder binary.ByteOrder) *sampleReader {
	if waveFmt.BitsPerSample == 8 {
		return &sampleReader{
			waveFmt:        waveFmt,
			byteOrder:      byteOrder,
			bytesPerSample: waveFmt.BitsPerSample / 8,
			divisor:        math.MaxUint8,
			midpoint:       128,
		}
	} else {
		return &sampleReader{
			waveFmt:        waveFmt,
			byteOrder:      byteOrder,
			bytesPerSample: waveFmt.BitsPerSample / 8,
			divisor:        math.Pow(2, float64(waveFmt.BitsPerSample-1)) - 1,
			midpoint:       0,
		}
	}
}

func (sr *sampleReader) readSample(sample []byte) (float64, error) {
	if len(sample) != sr.bytesPerSample {
		return 0, errors.New(fmt.Sprintf("expected %d bytes, actual %d bytes", sr.bytesPerSample, len(sample)))
	}
	paddedBytesPerSample := sr.bytesPerSample
	if paddedBytesPerSample > 4 && paddedBytesPerSample < 8 {
		paddedBytesPerSample = 8
	} else if paddedBytesPerSample > 2 && paddedBytesPerSample < 4 {
		paddedBytesPerSample = 4
	}
	paddedSample := sample
	if paddedBytesPerSample != sr.bytesPerSample {
		isNegative := false
		if sr.byteOrder == binary.BigEndian {
			isNegative = (sample[0] & 0x80) == 0x80
		} else {
			isNegative = (sample[len(sample)-1] & 0x80) == 0x80
		}
		paddedSample = make([]byte, paddedBytesPerSample)
		for i := range paddedSample {
			if isNegative {
				paddedSample[i] = 0xff
			} else {
				paddedSample[i] = 0
			}
		}
		if sr.byteOrder == binary.LittleEndian {
			for i := range sample {
				paddedSample[i] = sample[i]
			}
		} else {
			for i := range sample {
				paddedSample[i+paddedBytesPerSample-sr.bytesPerSample] = sample[i]
			}
		}
	}
	byteBuffer := bytes.NewBuffer(paddedSample)
	if paddedBytesPerSample == 1 {
		var v uint8
		err := binary.Read(byteBuffer, sr.byteOrder, &v)
		if err != nil {
			return 0, err
		}
		return (float64(v) - sr.midpoint) / sr.divisor, nil
	} else if paddedBytesPerSample == 2 {
		var v int16
		err := binary.Read(byteBuffer, sr.byteOrder, &v)
		if err != nil {
			return 0, err
		}
		return (float64(v) - sr.midpoint) / sr.divisor, nil
	} else if paddedBytesPerSample == 4 {
		var v int32
		err := binary.Read(byteBuffer, sr.byteOrder, &v)
		if err != nil {
			return 0, err
		}
		return (float64(v) - sr.midpoint) / sr.divisor, nil
	} else if paddedBytesPerSample == 8 {
		var v int64
		err := binary.Read(byteBuffer, sr.byteOrder, &v)
		if err != nil {
			return 0, err
		}
		return (float64(v) - sr.midpoint) / sr.divisor, nil
	} else {
		return 0, errors.New(fmt.Sprintf("unsupported bits per sample: %d", sr.waveFmt.BitsPerSample))
	}
}

func readBytes(reader io.Reader, len int) ([]byte, error) {
	data := make([]byte, len)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
