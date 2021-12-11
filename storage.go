package main

import (
	"bytes"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	PutFile(localPath string, destPath string) error
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

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func (s *S3Storage) PutFile(localPath string, destPath string) error {
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
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	return err
}
