package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Configuration struct {
	S3BucketName   string
	S3BucketRegion string
	S3MediaFolder  string
	S3BaseUrl      string
}

func GetConfig() *Configuration {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Print(err)
	}

	return &Configuration{
		S3BucketName:   os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion: os.Getenv("S3_BUCKET_REGION"),
		S3MediaFolder:  os.Getenv("S3_BUCKET_MEDIA_FOLDER"),
		S3BaseUrl:      os.Getenv("S3_BUCKET_PUBLIC_BASE_URL"),
	}
}
