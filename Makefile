.PHONY: all deploy clean

all: cert-bot cert-bot.zip

deploy: cert-bot.zip
	npm install
	cdk bootstrap
	cdk deploy

cert-bot.zip: cert-bot
	build-lambda-zip cert-bot

cert-bot: *.go
	export GOOS=linux
	go build

clean:
	rm -rf cert-bot cert-bot.zip
	rm -rf cdk.out node_modules package-lock.json
