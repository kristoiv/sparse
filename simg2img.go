package sparse

import "io"

func Simg2imgReader(input io.Reader) (io.Reader, error) {
	header, err := readHeader(input)
	if err != nil {
		return nil, err
	}
	return &sparseReader{input, header, nil, io.LimitReader(input, 0)}, nil
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
