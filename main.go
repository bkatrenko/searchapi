package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	elasticAddr    = "ELASTICSEARCH_URL"
	serverAddr     = "SERVER_ADDR"
	tokenSignKey   = "SIGN_KEY"
	tokenVerifyKey = "VERIFY_KEY"
	tokenExpireAt  = "TOKEN_EXPIRE_AT"
)

func main() {
	handlers, err := newAPIHandler(
		os.Getenv(elasticAddr),
		os.Getenv(tokenSignKey),
		os.Getenv(tokenVerifyKey),
		os.Getenv(tokenExpireAt))
	if err != nil {
		log.Fatal(err)
	}

	server := newServer(os.Getenv(serverAddr), handlers)

	go server.start()

	interruptSignal := make(chan os.Signal)
	signal.Notify(interruptSignal, os.Interrupt, syscall.SIGTERM)

	<-interruptSignal
	server.shutdown()
}
