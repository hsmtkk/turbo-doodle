#!/bin/sh
mkdir -p /data/test
mkdir -p /data/.minio.sys/config
mkdir -p /data/.minio.sys/buckets/test
cp /mnt/config.json /data/.minio.sys/config/config.json
cp /mnt/metadata.bin /data/.minio.sys/buckets/test/.metadata.bin
minio server /data --console-address :9001

