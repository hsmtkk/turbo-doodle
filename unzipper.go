package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/uuid"
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
	sugar.Info("connected minio")

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
	upBucket, err := newUploader(h.client, h.sugar).Upload(zipContent)
	if err != nil {
		h.sugar.Fatalf("failed to upload file", "error", err)
	}
	h.sugar.Infow("upload", "bucket", upBucket)
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

type uploader struct {
	client *minio.Client
	sugar  *zap.SugaredLogger
}

func newUploader(client *minio.Client, sugar *zap.SugaredLogger) *uploader {
	return &uploader{client, sugar}
}

func (u *uploader) Upload(zipContent []byte) (string, error) {
	bucket := uuid.New().String()
	if err := u.client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
		return "", fmt.Errorf("failed to make new bucket; %s; %w", bucket, err)
	}
	reader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		return "", fmt.Errorf("failed to open zip file; %w", err)
	}
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			u.sugar.Errorw("failed to open zip part", "file", f.Name, "error", err)
			continue
		}
		content, err := io.ReadAll(rc)
		if err != nil {
			u.sugar.Errorw("failed to read zip part", "file", f.Name, "error", err)
			continue
		}
		info, err := u.client.PutObject(context.Background(), bucket, f.Name, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
		if err != nil {
			u.sugar.Errorw("failed to put object", "file", f.Name, "error", err)
			continue
		}
		u.sugar.Infow("upload", "info", info)
	}
	return bucket, nil
}
