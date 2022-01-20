package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hsmtkk/turbo-doodle/env"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	natsHost := env.RequiredString("NATS_HOST")
	natsPort := env.RequiredInt("NATS_PORT")
	natsChannel := env.RequiredString("NATS_CHANNEL")
	natsURL := fmt.Sprintf("nats://%s:%d", natsHost, natsPort)

	minioHost := env.RequiredString("MINIO_HOST")
	minioPort := env.RequiredInt("MINIO_PORT")
	minioAccessKey := env.RequiredString("MINIO_ACCESS")
	minioSecretKey := env.RequiredString("MINIO_SECRET")
	minioAddress := fmt.Sprintf("%s:%d", minioHost, minioPort)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to init logger; %s", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		sugar.Fatalw("failed to connect NATS", "error", err)
	}
	defer natsConn.Close()
	sugar.Info("connected NATS")

	opts := minio.Options{
		Creds: credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
	}
	minioClient, err := minio.New(minioAddress, &opts)
	if err != nil {
		sugar.Fatalw("failed to connect minio", "error", err)
	}

	handler := newHandler(sugar, minioClient)
	sub, err := natsConn.Subscribe(natsChannel, handler.HandleMessage)
	if err != nil {
		log.Fatalf("failed to subscribe %s channel; %s", natsChannel, err)
	}
	defer sub.Unsubscribe()

	// block forever
	select {}
}

type handler struct {
	sugar  *zap.SugaredLogger
	client *minio.Client
}

func newHandler(sugar *zap.SugaredLogger, minioClient *minio.Client) *handler {
	return &handler{sugar, minioClient}
}

func (h *handler) HandleMessage(msg *nats.Msg) {
	h.sugar.Debugw("handle message", "msg", string(msg.Data))
	ev := Event{}
	if err := json.Unmarshal(msg.Data, &ev); err != nil {
		h.sugar.Fatalf("failed to unmarshal JSON", "msg", string(msg.Data), "error", err)
	}
	elems := strings.Split(ev.Key, "/")
	bucket := elems[0]
	fileName := elems[1]
	h.sugar.Infow("bucket and file", "bucket", bucket, "file", fileName)
	zipContent, err := newDownloader(h.client, bucket).Download(fileName)
	if err != nil {
		h.sugar.Fatalf("failed to download zip", "bucket", bucket, "file", fileName, "error", err)
	}
	h.sugar.Info(len(zipContent))
}

type Event struct {
	Key string
}

type downloader struct {
	client *minio.Client
	bucket string
}

func newDownloader(client *minio.Client, bucket string) *downloader {
	return &downloader{client, bucket}
}

func (d *downloader) Download(fileName string) ([]byte, error) {
	obj, err := d.client.GetObject(context.Background(), d.bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object; %w", err)
	}
	body, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object; %w", err)
	}
	return body, nil
}
