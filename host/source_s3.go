package host

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"os"
)

const S3ConfigKey = "gaffer.json"

type S3Source struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
	key        string
}

func (ss S3Source) Get() (*Config, error) {
	fp, err := ioutil.TempFile("/tmp", "")
	if err != nil {
		return nil, err
	}
	defer os.Remove(fp.Name())
	_, err = ss.downloader.Download(fp, &s3.GetObjectInput{
		Bucket: &ss.bucket,
		Key:    &ss.key,
	})
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.NewDecoder(fp).Decode(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (ss S3Source) Set(config *Config) error {
	fp, err := ioutil.TempFile("/tmp", "")
	if err != nil {
		return err
	}
	defer os.Remove(fp.Name())
	err = json.NewEncoder(fp).Encode(config)
	if err != nil {
		return err
	}
	_, err = ss.uploader.Upload(&s3manager.UploadInput{
		Bucket: &ss.bucket,
		Key:    &ss.key,
		Body:   fp,
	})
	return err
}

func NewS3Source(bucket, key string) (*S3Source, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return &S3Source{
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     bucket,
		key:        key,
	}, nil
}
