package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	opts := minio.Options{
		Creds: credentials.NewStaticV4("hogehoge", "fugafuga", ""),
	}
	client, err := minio.New("192.168.11.13:9000", &opts)
	if err != nil {
		log.Fatalf("failed to connect minio; %s", err)
	}
	zipBytes, err := download(client)
	if err != nil {
		log.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir("tmp", "")
	if err != nil {
		log.Fatalf("failed to create temporary directory; %s", err)
	}
	defer os.RemoveAll(tmpDir)

	extractZip(tmpDir, zipBytes)

	upload(tmpDir, client)
}

func download(client *minio.Client) ([]byte, error) {
	obj, err := client.GetObject(context.Background(), "test", "sample.zip", minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object; %w", err)
	}
	body, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object; %w", err)
	}
	return body, nil
}

func extractZip(tmpDir string, zipBytes []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("failed to open zip file; %w", err)
	}
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			log.Printf("failed to open file in zip; %s; %s", f.Name, err)
			continue
		}
		outPath := filepath.Join(tmpDir, f.Name)
		out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("failed to open file to write; %s; %s", outPath, err)
			continue
		}
		if _, err := io.Copy(out, rc); err != nil {
			log.Printf("failed to copy; %s", err)
		}
		log.Printf("extracted; %s", outPath)
	}
	return nil
}

func upload(tmpDir string, client *minio.Client) error {
	bucket := uuid.New().String()
	if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("failed to make bucket; %s", err)
	}
	return nil

}
