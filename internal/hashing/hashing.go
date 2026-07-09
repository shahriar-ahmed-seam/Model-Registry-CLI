package hashing

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

const chunkSize = 1 << 20 // 1 MiB

// HashReader computes the lowercase hex SHA256 over all bytes of r.
// Returns the hex digest, the number of bytes read, and any error encountered.
func HashReader(r io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.CopyBuffer(h, r, make([]byte, chunkSize))
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// HashFile opens path and streams it through HashReader.
// Returns the hex digest, file size in bytes, and any error encountered.
func HashFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	
	// Use buffered reader with chunkSize buffer
	br := bufio.NewReaderSize(f, chunkSize)
	return HashReader(br)
}