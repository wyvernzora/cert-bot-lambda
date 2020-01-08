package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type CredentialsManager struct {
	server         string
	secretsManager *secretsmanager.SecretsManager
}

func NewCredentialsManager(sess *session.Session, server string) *CredentialsManager {
	return &CredentialsManager{
		server:         server,
		secretsManager: secretsmanager.New(sess),
	}
}

func (credentials *CredentialsManager) CreateOrRetrieve() (crypto.PrivateKey, error) {
	privateKey, err := credentials.Retrieve()
	if err != nil {
		return nil, err
	}
	if privateKey != nil {
		return privateKey, nil
	}
	return credentials.Create()
}

func (credentials *CredentialsManager) Create() (crypto.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	pem, err := encodePrivateKey(key)
	if err != nil {
		return nil, err
	}

	createSecretInput := &secretsmanager.CreateSecretInput{
		Name:         generateCredentialName(credentials.server),
		SecretString: &pem,
	}
	if _, err = credentials.secretsManager.CreateSecret(createSecretInput); err != nil {
		return nil, err
	}

	return key, nil
}

func (credentials *CredentialsManager) Retrieve() (crypto.PrivateKey, error) {
	getSecretInput := &secretsmanager.GetSecretValueInput{
		SecretId: generateCredentialName(credentials.server),
	}
	output, err := credentials.secretsManager.GetSecretValue(getSecretInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				return nil, nil
			default:
				return nil, err
			}
		}
	}

	key, err := decodePrivateKey(*output.SecretString)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func generateCredentialName(server string) *string {
	secretID := fmt.Sprintf("acme/%s", server)
	return &secretID
}

func encodePrivateKey(key crypto.PrivateKey) (string, error) {
	pkcs, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}

	pem := pem.EncodeToMemory(
		&pem.Block{
			Type: "ECDSA PRIVATE KEY",

			Bytes: pkcs,
		},
	)

	return string(pem), nil
}

func decodePrivateKey(blob string) (crypto.PrivateKey, error) {
	block, _ := pem.Decode([]byte(blob))
	if block == nil {
		return nil, errors.New("Could not parse private key block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}
