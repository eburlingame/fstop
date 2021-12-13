package main

import (
	"bytes"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	PutFile(localPath string, destPath string, contentType string) error
	GetSignedUploadUrl(destPath string, contentType string) (string, error)
}

type S3Storage struct {
	session    *session.Session
	bucketName string
}

func InitS3Storage(config *Configuration) (*S3Storage, error) {
	aws, err := session.NewSession(&aws.Config{
		Region: aws.String(config.S3BucketRegion),
	})

	if err != nil {
		return nil, err
	}

	storage := &S3Storage{
		session:    aws,
		bucketName: config.S3BucketName,
	}

	return storage, nil
}

func (s *S3Storage) PutFile(localPath string, destPath string, contentType string) error {
	// Open the file for use
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s.session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.bucketName),
		Key:                  aws.String(destPath),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(contentType),
		ServerSideEncryption: aws.String("AES256"),
	})

	return err
}

func (s *S3Storage) GetSignedUploadUrl(destPath string, contentType string) (string, error) {
	svc := s3.New(s.session)

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(destPath),
		ContentType: aws.String(contentType),
	})
	urlStr, err := req.Presign(15 * time.Minute)

	return urlStr, err
}
