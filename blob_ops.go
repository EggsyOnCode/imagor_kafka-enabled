package imagor

import (
	"bytes"
	"io"
	"net/http"
)

// readSeekCloser wraps a bytes.Reader to implement io.ReadSeekCloser
type readSeekCloser struct {
	*bytes.Reader
}

// Close is a no-op for readSeekCloser, since bytes.Reader doesn't require cleanup
func (rsc readSeekCloser) Close() error {
	return nil
}

// BytesToBlob converts a byte array to a Blob
func BytesToBlob(data []byte, contentType string) *Blob {
	size := int64(len(data))

	return &Blob{
		newReader: func() (io.ReadCloser, int64, error) {
			return io.NopCloser(bytes.NewReader(data)), size, nil
		},
		newReadSeeker: func() (io.ReadSeekCloser, int64, error) {
			return readSeekCloser{bytes.NewReader(data)}, size, nil
		},
		sniffBuf:    data[:min(len(data), 512)],                         // Use the first 512 bytes for sniffing content type
		size:        size,                                               // Set the size of the Blob
		contentType: contentType,                                        // Set the content type for the Blob
		Header:      http.Header{"Content-Type": []string{contentType}}, // Set HTTP header
	}
}

// Helper function to ensure we don't slice out of bounds
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
