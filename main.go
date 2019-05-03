package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	envFile = "app.env"

	elasticAddr    = "ELASTIC_ADDR"
	serverAddr     = "SERVER_ADDR"
	tokenSignKey   = "SIGN_KEY"
	tokenVerifyKey = "VERIFY_KEY"
	tokenExpireAt  = "TOKEN_EXPIRE_AT"
)

func main() {
	godotenv.Load(envFile)

	handlers, err := newAPIHandler(
		os.Getenv(elasticAddr),
		os.Getenv(tokenSignKey),
		os.Getenv(tokenVerifyKey),
		os.Getenv(tokenExpireAt))
	if err != nil {
		log.Fatal(err)
	}

	server := newServer(os.Getenv(serverAddr), handlers)

	interruptSignal := make(chan os.Signal)
	done := make(chan bool)
	signal.Notify(interruptSignal, os.Interrupt, syscall.SIGTERM)

	go server.start()

	go func() {
		<-interruptSignal
		server.shutdown()
		done <- true
	}()

	<-done
}
