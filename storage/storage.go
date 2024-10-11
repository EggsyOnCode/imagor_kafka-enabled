package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// BlobFile implements multipart.File interface
type BlobFile struct {
	*bytes.Reader
}

type GCPBucketConfig struct {
	ProjectID  string
	BucketName string
}

// Ensure BlobFile satisfies the multipart.File interface
var _ multipart.File = (*BlobFile)(nil)

type ClientUploader struct {
	cl         *storage.Client
	ProjectID  string
	BucketName string
	UploadPath string
}

func NewClientUploader(pId string, bName string) (*ClientUploader, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx) // Create the Google Cloud Storage client
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", err)
	}

	// Initialize the ClientUploader struct
	uploader := &ClientUploader{
		cl:         client,
		ProjectID:  pId,
		BucketName: bName,
		UploadPath: "",
	}

	return uploader, nil
}

// UploadFile uploads an object
func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	log.Printf("Uploading file to bucket...")

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	log.Printf("File %v uploaded to %v\n", object, c.BucketName)

	return nil
}

// GetFile downloads an object from the bucket and writes it to a destination file
func (c *ClientUploader) GetFile(object, destFileName string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Open the destination file for writing
	f, err := os.Create(destFileName)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer f.Close()

	// Create a new reader for the object in the bucket
	rc, err := c.cl.Bucket(c.BucketName).Object(c.UploadPath + object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", object, err)
	}
	defer rc.Close()

	// Copy the object data to the local file
	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	fmt.Printf("File %v downloaded to %v\n", object, destFileName)
	return nil
}

// Close is part of the multipart.File interface
func (b *BlobFile) Close() error {
	// You can implement any cleanup logic here if needed
	return nil
}

// ConvertBlobToMultipartFile converts a byte slice (Blob) to multipart.File
func ConvertBlobToMultipartFile(blob []byte) multipart.File {
	// Use bytes.Reader to treat the byte slice like a file
	return &BlobFile{Reader: bytes.NewReader(blob)}
}
