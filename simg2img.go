package sparse

import (
	"bytes"
	"errors"
	"io"
)

func Simg2imgReader(input io.Reader) (io.Reader, error) {
	header, err := readHeader(input)
	if err != nil {
		return nil, err
	}
	return &sparseReader{input, header, nil, io.LimitReader(input, 0)}, nil
}

func Simg2imgWriter(output io.WriteSeeker) io.Writer {
	return &sparseWriter{output, 0, new(bytes.Buffer), nil, nil, 0, 28, 0}
}

type sparseReader struct {
	io.Reader
	header             *SparseHeader
	currentChunkHeader *ChunkHeader
	currentChunkReader io.Reader
}

func (self *sparseReader) Read(p []byte) (n int, err error) {
	bytesRead := 0
	for {
		n, err := self.currentChunkReader.Read(p[bytesRead:])
		bytesRead = bytesRead + n
		if err != nil && err != io.EOF {
			return bytesRead, err
		} else if err == io.EOF {
			// Load next chunk (will run on the first read since initial self.currentChunkReader returns EOF on reads)
			chunkHeader, err := readChunkHeader(self.Reader)
			if err != nil && err != io.EOF {
				return bytesRead, err
			} else if err == io.EOF {
				// No more chunks
				return bytesRead, io.EOF
			}
			self.currentChunkHeader = chunkHeader
			chunkReader, err := chunkReader(self.header, self.currentChunkHeader, self.Reader)
			if err != nil {
				return bytesRead, err
			}
			self.currentChunkReader = chunkReader
		} else {
			if bytesRead == len(p) {
				// We are done
				break
			} else {
				// Shorter read than expected. Lets try reading more to see what happens.
				// (the current chunk is probably empty, must be replaced with a new one)
			}
		}
	}

	return bytesRead, nil
}

type state int

const (
	state_reading_header state = iota
	state_reading_chunk_header
	state_reading_raw
	state_failed
)

type sparseWriter struct {
	output                io.WriteSeeker
	state                 state
	buffer                *bytes.Buffer
	header                *SparseHeader
	currentChunkHeader    *ChunkHeader
	currentChunkBytesRead int64
	totalSize             int64
	chunkIndex            int
}

func (self *sparseWriter) Write(p []byte) (n int, err error) {
	if self.state == state_failed {
		return 0, errors.New("Writing to failed simg2img writer")
	}

	n, err = self.buffer.Write(p)
	if err != nil {
		return
	}

	for {
		if self.buffer.Len() == 0 {
			return
		}

		switch self.state {

		case state_reading_header:
			if self.buffer.Len() >= 28 {
				header, er := readHeader(self.buffer)
				if er != nil {
					err = er
					return
				}
				self.header = header
				self.state = state_reading_chunk_header
			} else {
				return
			}

		case state_reading_chunk_header:
			if self.buffer.Len() >= 12 {
				chunkHeader, er := readChunkHeader(self.buffer)
				if er != nil {
					err = er
					return
				}

				self.chunkIndex += 1
				self.currentChunkHeader = chunkHeader
				self.currentChunkBytesRead = 0
				if chunkHeader.ChunkType == type_raw {
					self.state = state_reading_raw
				} else if chunkHeader.ChunkType == type_dont_care {
					l := self.header.BlockSize * chunkHeader.ChunkSize
					self.output.Seek(int64(l), io.SeekCurrent)
					if self.chunkIndex == int(self.header.TotalChunks) {
						self.output.Seek(-1, io.SeekCurrent)
						nw, er := self.output.Write([]byte{0x00})
						if er != nil {
							err = er
							return
						} else if nw != 1 {
							err = errors.New("Unable to write ending 0x00 byte")
							return
						}
						return // Aaaand we're done!
					}
					self.state = state_reading_chunk_header
				} else { // Unsupported TYPE_FILL
					err = errors.New("Unsupported chunk type type_fill")
					self.state = state_failed
					return
				}
			} else {
				return
			}

		case state_reading_raw:
			l := int64(self.header.BlockSize*self.currentChunkHeader.ChunkSize) - self.currentChunkBytesRead
			nw, er := io.Copy(self.output, io.LimitReader(self.buffer, l))
			self.currentChunkBytesRead += nw
			if er != nil && er != io.EOF {
				err = er
				self.state = state_failed
				return
			}
			if nw == l {
				// We're done with this chunk
				self.state = state_reading_chunk_header
			} else {
				// This was as far as we got with what we have in self.buffer
				return
			}

		}
	}
}
