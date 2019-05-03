package main

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type instance struct {
	addr     string
	handlers apiHandler

	httpServer *http.Server
}

func newServer(addr string, handlers apiHandler) *instance {
	s := &instance{
		addr:     addr,
		handlers: handlers,
	}

	return s
}

func (s *instance) start() {
	s.httpServer = &http.Server{Addr: s.addr, Handler: s.handlers.router}
	err := s.httpServer.ListenAndServe()

	if err != http.ErrServerClosed {
		logrus.WithError(err).Error("http Server stopped unexpected")
		s.shutdown()
	} else {
		logrus.WithError(err).Info("http Server stopped")
	}
}

func (s *instance) shutdown() {
	if s.httpServer != nil {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			logrus.WithError(err).Error("Failed to shutdown http server gracefully")
		} else {
			s.httpServer = nil
		}
	}
}
