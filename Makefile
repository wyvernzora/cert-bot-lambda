.PHONY: all clean

all: cert-bot cert-bot.zip

cert-bot.zip: cert-bot
	build-lambda-zip cert-bot

cert-bot: *.go
	export GOOS=linux
	go build

clean:
	rm -rf cert-bot cert-bot.zip
