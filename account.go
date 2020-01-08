package main

import (
	"crypto"

	"github.com/go-acme/lego/registration"
)

// Account is the data type storing credentials for the given ACME registration.
// Persisting account data allows reuse of past authorizations for renewal of certificates.
type Account struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

// New creates an account instance from email and credentials
func NewAccount(email string, key crypto.PrivateKey) *Account {
	return &Account{
		Email: email,
		Key:   key,
	}
}

// GetEmail returns the email address associated with the account
func (u *Account) GetEmail() string {
	return u.Email
}

// GetRegistration returns ACME account info as provided by the registrar
func (u Account) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the credentials to the ACME account
func (u *Account) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}
