package resources

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	PutFile(contents []byte, destPath string, contentType string) error
	GetSignedUploadUrl(destPath string, contentType string) (string, error)
	ListFiles(prefix string) ([]string, error)
	GetFile(key string) ([]byte, error)
	DeleteFile(key string) error
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

func (s *S3Storage) PutFile(contents []byte, destPath string, contentType string) error {
	size := int64(len(contents))

	_, err := s3.New(s.session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.bucketName),
		Key:                  aws.String(destPath),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(contents),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(contentType),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		log.Printf("Error uploading file: %s", err)
	}

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

func (s *S3Storage) ListFiles(prefix string) ([]string, error) {
	svc := s3.New(s.session)

	list, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, object := range list.Contents {
		// Filter out directories
		if !strings.HasSuffix(*object.Key, "/") {
			names = append(names, *object.Key)
		}
	}

	return names, nil
}

func (s *S3Storage) GetFile(key string) ([]byte, error) {
	svc := s3.New(s.session)

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(obj.Body)
}

func (s *S3Storage) DeleteFile(key string) error {
	svc := s3.New(s.session)

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	return err
}
