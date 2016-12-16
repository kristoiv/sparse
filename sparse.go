package sparse

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type SparseHeader struct {
	Magic           uint32 // 0xed26ff3a
	MajorVersion    uint16 // (0x1) - reject images with higher major versions
	MinorVersion    uint16 // (0x0) - allow images with higer minor versions
	FileHeaderSize  uint16 // 28 bytes for first revision of the file format
	ChunkHeaderSize uint16 // 12 bytes for first revision of the file format
	BlockSize       uint32 // block size in bytes, must be a multiple of 4 (4096)
	TotalBlocks     uint32 // total blocks in the non-sparse output image
	TotalChunks     uint32 // total chunks in the sparse input image
	ImageChecksum   uint32 // CRC32 checksum of the original data, counting "don't care"
}

type ChunkHeader struct {
	ChunkType uint16 // 0xCAC1 -> raw; 0xCAC2 -> fill; 0xCAC3 -> don't care
	Reserved  uint16
	ChunkSize uint32 // in blocks in output image
	TotalSize uint32 // in bytes of chunk input file including chunk header and data
}

const (
	type_raw       = 0xcac1
	type_fill      = 0xcac2
	type_dont_care = 0xcac3
)

func readHeader(r io.Reader) (*SparseHeader, error) {
	buf := make([]byte, 28)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	h := &SparseHeader{}
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, h); err != nil {
		return nil, err
	}

	if h.Magic != uint32(0xed26ff3a) {
		return nil, fmt.Errorf("Not a valid sparse file. Magic byte value was really 0x%x", h.Magic)
	}

	if h.MajorVersion != 0x1 {
		return nil, fmt.Errorf("Sparse file not of a supported major version. Actually: 0x%x", h.MajorVersion)
	}

	if h.MinorVersion != 0x0 {
		return nil, fmt.Errorf("Sparse file not of a supported major version. Actually: 0x%x", h.MinorVersion)
	}

	if h.FileHeaderSize != 28 {
		return nil, fmt.Errorf("Sparse file contains unsupported fileHeaderSize (actual=%d supported=%d)", h.FileHeaderSize, 28)
	}

	if h.ChunkHeaderSize != 12 {
		return nil, fmt.Errorf("Sparse file contains unsupported chunkHeaderSize (actual=%d supported=%d)", h.ChunkHeaderSize, 12)
	}

	return h, nil
}

func readChunkHeader(r io.Reader) (*ChunkHeader, error) {
	buf := make([]byte, 12)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}

	h := &ChunkHeader{}
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, h); err != nil {
		return nil, err
	}

	if h.ChunkType < type_raw || h.ChunkType > type_dont_care {
		return nil, fmt.Errorf("Sparse file contains a chunk with unsupported chunk type (0x%x)", h.ChunkType)
	}

	return h, nil
}

func chunkReader(sparseHeader *SparseHeader, chunkHeader *ChunkHeader, r io.Reader) (io.Reader, error) {
	if chunkHeader.ChunkType == type_fill {
		return nil, errors.New("Error reading sparse file. Unsupported chunk type type_fill")
	}

	var out io.Reader
	if chunkHeader.ChunkType == type_raw {
		// We use the underlying file
		out = r
	} else {
		// We dont care, so we fill with random noize
		out = &dontCareReader{}
	}

	l := sparseHeader.BlockSize * chunkHeader.ChunkSize
	return io.LimitReader(out, int64(l)), nil
}

type dontCareReader struct{}

func (self *dontCareReader) Read(p []byte) (n int, err error) {
	return len(p), nil
}
