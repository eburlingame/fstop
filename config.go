package main

import "os"

type Configuration struct {
	S3BucketName   string
	S3BucketRegion string
}

func GetConfig() *Configuration {
	return &Configuration{
		S3BucketName:   os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion: os.Getenv("S3_BUCKET_REGION"),
	}
}
