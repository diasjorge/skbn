package skbn

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetClientToS3 checks the connection to S3 and returns the tested client
func GetClientToS3() (*session.Session, error) {
	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		s, err := getNewSession()
		if err != nil {
			if attempt == attempts {
				return nil, err
			}
			time.Sleep(1 * time.Second)
			continue
		}

		_, err = s3.New(s).ListBuckets(&s3.ListBucketsInput{})
		if attempt == attempts {
			if err != nil {
				return nil, err
			}
		}
		if err == nil {
			return s, nil
		}
		time.Sleep(1 * time.Second)
	}

	return nil, nil
}

// GetListOfFilesFromS3 gets list of files in path from S3 (recursive)
func GetListOfFilesFromS3(s *session.Session, path string) ([]string, error) {
	pSplit := strings.Split(path, "/")
	if err := validateS3Path(pSplit); err != nil {
		return nil, err
	}
	bucket := pSplit[0]
	pathToCopy := filepath.Join(pSplit[1:]...)

	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		objectOutput, err := s3.New(s).ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String(pathToCopy),
		})
		if err != nil {
			if attempt == attempts {
				return nil, err
			}
			time.Sleep(1 * time.Second)
			continue
		}

		var outLines []string
		for _, content := range objectOutput.Contents {
			line := *content.Key
			outLines = append(outLines, strings.Replace(line, pathToCopy, "", 1))
		}

		return outLines, nil
	}

	return nil, nil
}

// DownloadFromS3 downloads a single file from S3
func DownloadFromS3(s *session.Session, path string) ([]byte, error) {
	pSplit := strings.Split(path, "/")
	if err := validateS3Path(pSplit); err != nil {
		return nil, err
	}
	bucket := pSplit[0]
	s3Path := filepath.Join(pSplit[1:]...)

	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		objectOutput, err := s3.New(s).GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Path),
		})
		if err != nil {
			if attempt == attempts {
				return nil, err
			}
			time.Sleep(1 * time.Second)
			continue
		}

		buffer := make([]byte, int(*objectOutput.ContentLength))
		objectOutput.Body.Read(buffer)
		return buffer, nil
	}

	return nil, nil
}

// UploadToS3 uploads a single file to S3
func UploadToS3(s *session.Session, toPath, fromPath string, buffer []byte) error {
	pSplit := strings.Split(toPath, "/")
	if err := validateS3Path(pSplit); err != nil {
		return err
	}
	if len(pSplit) == 1 {
		_, fileName := filepath.Split(fromPath)
		pSplit = append(pSplit, fileName)
	}
	bucket := pSplit[0]
	s3Path := filepath.Join(pSplit[1:]...)

	attempts := 3
	attempt := 0
	for attempt < attempts {
		attempt++

		_, err := s3.New(s).PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Path),
			Body:   bytes.NewReader(buffer),
		})
		if attempt == attempts {
			if err != nil {
				return err
			}
		}
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func validateS3Path(pathSplit []string) error {
	if len(pathSplit) >= 1 {
		return nil
	}
	return fmt.Errorf("illegal path: %s", filepath.Join(pathSplit...))
}

func getNewSession() (*session.Session, error) {
	region := "eu-central-1"
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	return s, err
}
