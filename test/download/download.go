package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	opts := minio.Options{
		Creds: credentials.NewStaticV4("hogehoge", "fugafuga", ""),
	}
	minioClient, err := minio.New("192.168.11.13:9000", &opts)
	if err != nil {
		log.Fatalf("failed to connect minio; %s", err)
	}
	obj, err := minioClient.GetObject(context.Background(), "test", "sample.txt", minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("failed to get object; %s", err)
	}
	body, err := io.ReadAll(obj)
	if err != nil {
		log.Fatalf("failed to read object; %s", err)
	}
	fmt.Println(string(body))
}
