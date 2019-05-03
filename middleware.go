package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Middleware function, which will be called for each request
func (h *apiHandler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth":
			next.ServeHTTP(w, r)
		default:
			token := r.Header.Get("Authorization")

			claims, err := verifyToken(token, h.verifyKey)
			if err != nil {
				log.Errorf("error while verify token: %v", err)
				h.writeError(w, http.StatusUnauthorized, "bad token")
				return
			}

			log.Infof("auth new user success, token source is: %s", claims.Info.Source)
			next.ServeHTTP(w, r)
		}
	})
}

// Middleware function, which will be called for each request
func (h *apiHandler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s %v", r.Method, r.URL.Path, r.URL.Query())
		next.ServeHTTP(w, r)
	})
}
