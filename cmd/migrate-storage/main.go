package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	dstEndpoint := os.Getenv("R2_ENDPOINT")
	dstAccessKey := os.Getenv("R2_ACCESS_KEY")
	dstSecretKey := os.Getenv("R2_SECRET_KEY")
	dstBucket := os.Getenv("R2_BUCKET")

	if dstEndpoint == "" || dstAccessKey == "" || dstSecretKey == "" || dstBucket == "" {
		log.Fatal("Set R2_ENDPOINT, R2_ACCESS_KEY, R2_SECRET_KEY, R2_BUCKET")
	}

	ctx := context.Background()

	src, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("pisaupedia_admin", "pisaupedia_minio_secret", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatal("src connect:", err)
	}

	dst, err := minio.New(dstEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(dstAccessKey, dstSecretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal("dst connect:", err)
	}

	srcBucket := "pisaupedia"
	objects := src.ListObjects(ctx, srcBucket, minio.ListObjectsOptions{Recursive: true})

	count := 0
	for obj := range objects {
		if obj.Err != nil {
			log.Printf("list error: %v", obj.Err)
			continue
		}

		reader, err := src.GetObject(ctx, srcBucket, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("GET %s: %v", obj.Key, err)
			continue
		}

		_, err = dst.PutObject(ctx, dstBucket, obj.Key, reader, obj.Size, minio.PutObjectOptions{})
		reader.Close()

		if err != nil {
			log.Printf("PUT %s: %v", obj.Key, err)
			continue
		}
		count++
		fmt.Printf("OK %s (%d bytes)\n", obj.Key, obj.Size)
	}

	fmt.Printf("\nDone! %d files migrated.\n", count)
}
