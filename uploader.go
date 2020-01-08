package main

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-acme/lego/certificate"
)

type Uploader struct {
	S3     *s3.S3
	Bucket string
}

func NewUploader(sess *session.Session, bucket string) *Uploader {
	return &Uploader{
		S3:     s3.New(sess),
		Bucket: bucket,
	}
}

func (uploader *Uploader) UploadCertificate(domain string, certs *certificate.Resource) error {
	certPath := fmt.Sprintf("%s/%s.crt", domain, domain)
	if err := uploader.uploadObject(certPath, certs.Certificate); err != nil {
		return err
	}

	keyPath := fmt.Sprintf("%s/%s.key", domain, domain)
	if err := uploader.uploadObject(keyPath, certs.PrivateKey); err != nil {
		return err
	}

	caPath := fmt.Sprintf("%s/%s.ca.crt", domain, domain)
	if err := uploader.uploadObject(caPath, certs.IssuerCertificate); err != nil {
		return err
	}

	return nil
}

func (uploader *Uploader) uploadObject(path string, payload []byte) error {
	putObjectInput := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(payload)),
		Bucket: aws.String(uploader.Bucket),
		Key:    aws.String(path),
	}
	if _, err := uploader.S3.PutObject(putObjectInput); err != nil {
		return err
	}
	return nil
}
