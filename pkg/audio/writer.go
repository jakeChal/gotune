package audio

import (
	"encoding/binary"
	"io"
	"os"
)

type Writer struct {
	file         *os.File
	sampleRate   uint32
	channels     uint32
	bytesWritten uint32
}

func NewWriter(filename string, sampleRate, channels uint32) (*Writer, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := &Writer{
		file:       f,
		sampleRate: sampleRate,
		channels:   channels,
	}

	bitsPerSample := uint16(16)
	bytesPerSample := uint16(2)
	byteRate := sampleRate * uint32(channels) * uint32(bytesPerSample)
	blockAlign := uint16(channels) * bytesPerSample

	write := func(data any) error {
		switch v := data.(type) {
		case string:
			_, err := f.Write([]byte(v))
			return err
		default:
			return binary.Write(f, binary.LittleEndian, data)
		}
	}

	// Write RIFF header (12 bytes): "RIFF" + fileSize(0) + "WAVE"
	if err := write("RIFF"); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(uint32(0)); err != nil {
		f.Close()
		return nil, err
	}
	if err := write("WAVE"); err != nil {
		f.Close()
		return nil, err
	}

	// Write fmt chunk (24 bytes): format info (PCM, sample rate, channels, bit depth)
	if err := write("fmt "); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(uint32(16)); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(uint16(1)); err != nil { // PCM format
		f.Close()
		return nil, err
	}
	if err := write(uint16(channels)); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(sampleRate); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(byteRate); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(blockAlign); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(bitsPerSample); err != nil {
		f.Close()
		return nil, err
	}

	// Write data chunk header (8 bytes): "data" + dataSize(0)
	if err := write("data"); err != nil {
		f.Close()
		return nil, err
	}
	if err := write(uint32(0)); err != nil {
		f.Close()
		return nil, err
	}

	return writer, nil
}

func (w *Writer) WriteFrame(samples []float32) error {
	for _, sample := range samples {
		// Clamp sample to [-1.0, 1.0] range
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}

		// Convert to int16 PCM
		pcmValue := int16(sample * 32767)

		// Write to file
		if err := binary.Write(w.file, binary.LittleEndian, pcmValue); err != nil {
			return err
		}

		// Track bytes written (2 bytes per int16 sample)
		w.bytesWritten += 2
	}
	return nil
}

func (w *Writer) Close() error {
	// Seek back to byte 4 in the file (after "RIFF")
	if _, err := w.file.Seek(4, io.SeekStart); err != nil {
		return err
	}

	// Write the total file size minus 8 bytes (bytesWritten + 36)
	totalFileSize := w.bytesWritten + 36
	if err := binary.Write(w.file, binary.LittleEndian, totalFileSize); err != nil {
		return err
	}

	// Seek to byte 40 (after "data" chunk header)
	if _, err := w.file.Seek(40, io.SeekStart); err != nil {
		return err
	}
	if err := binary.Write(w.file, binary.LittleEndian, w.bytesWritten); err != nil {
		return err
	}
	return w.file.Close()
}
