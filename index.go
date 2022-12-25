package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/providers/dns/route53"
	"github.com/go-acme/lego/registration"
)

func main() {
	lambda.Start(requestCertificate)
}

var server, _ = os.LookupEnv("ACME_SERVER")
var bucket, _ = os.LookupEnv("OUTPUT_BUCKET")
var email, _ = os.LookupEnv("ACCOUNT_EMAIL")
var domain, _ = os.LookupEnv("FQDN")

func requestCertificate() {

	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String("us-west-2"),
		},
	))
	credentials := NewCredentialsManager(sess, server)
	uploader := NewUploader(sess, bucket)

	key, err := credentials.CreateOrRetrieve()
	if err != nil {
		log.Fatal(err)
	}
	account := NewAccount(email, key)

	config := lego.NewConfig(account)
	config.CADirURL = fmt.Sprintf("https://%s/directory", server)
	config.Certificate.KeyType = certcrypto.EC384

	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	registerOptions := registration.RegisterOptions{
		TermsOfServiceAgreed: true,
	}
	registration, err := client.Registration.Register(registerOptions)
	if err != nil {
		log.Fatal(err)
	}
	account.Registration = registration

	provider, err := route53.NewDNSProvider()
	if err != nil {
		log.Fatal(err)
	}
	client.Challenge.SetDNS01Provider(provider)

	request := certificate.ObtainRequest{
		Domains: []string{domain, "*." + domain},
		Bundle:  true,
	}
	certificate, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	if err := uploader.UploadCertificate(domain, certificate); err != nil {
		log.Fatal(err)
	}
	log.Println("Uploaded certificate")

}
