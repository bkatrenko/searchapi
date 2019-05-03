package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	wrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type apiHandler struct {
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey
	router    *mux.Router

	clientCreds   string
	tokenExpireAt int64

	elastic elasticEngine
}

func newAPIHandler(elasticAddr, privateKey, publicKey string, tokenExpireAt string) (apiHandler, error) {
	log.Info("connect to elastic on addr: ", elasticAddr)
	elastic, err := newElasticEngine(elasticAddr)
	if err != nil {
		return apiHandler{}, wrapper.Wrap(err, "can't create elastic engine")
	}

	expireAt, err := strconv.ParseInt(tokenExpireAt, 10, 64)
	if err != nil {
		return apiHandler{}, wrapper.Wrap(err, "error while parse expire at")
	}

	handlers := apiHandler{
		router:        mux.NewRouter(),
		elastic:       elastic,
		tokenExpireAt: expireAt,
	}

	if err := handlers.readKeys(privateKey, publicKey); err != nil {
		return apiHandler{}, wrapper.Wrap(err, "error while read token keys")
	}

	handlers.addV1Routes()
	return handlers, nil
}

func (h *apiHandler) addV1Routes() {
	r := h.router.PathPrefix("/v1").Subrouter()
	r.HandleFunc("/auth", h.getTokenV1).Methods("GET")
	r.HandleFunc("/products", h.getProductsV1).Methods("GET")
	r.HandleFunc("/products", h.postProductV1).Methods("POST")
	r.Use(h.authMiddleware, h.loggingMiddleware)
}

func (h *apiHandler) getTokenV1(w http.ResponseWriter, r *http.Request) {
	creds, ok := r.URL.Query()["creds"]
	if !ok || len(creds) != 1 {
		log.Error("can't get token: empty creds")
		h.writeError(w, http.StatusBadRequest, "bad request: empty creds")
		return
	}

	if creds[0] != os.Getenv("CLIENT_CREDS") {
		log.Info("can't get token: bad creds")
		h.writeError(w, http.StatusNotFound, "bad creds")
		return
	}

	token, err := makeToken(h.signKey, "searchapi", time.Now().Add(time.Duration(h.tokenExpireAt)*time.Second).Unix())
	if err != nil {
		log.Errorf("error while make new token: %v", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.writeJSON(w, http.StatusOK, Token{token})
}

func (h *apiHandler) getProductsV1(w http.ResponseWriter, r *http.Request) {
	params := queryParams{}
	if err := params.fill(r); err != nil {
		log.Error(err)
		h.writeError(w, http.StatusBadRequest, "bad query")
		return
	}

	if err := params.validate(); err != nil {
		log.Error(err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	docs, err := h.elastic.get(params)
	if err != nil {
		log.Error(err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := h.writeJSON(w, http.StatusOK, docs); err != nil {
		log.Error(err)
	}
}

func (h *apiHandler) postProductV1(w http.ResponseWriter, r *http.Request) {
	docs := []Doc{}

	if err := json.NewDecoder(r.Body).Decode(&docs); err != nil {
		log.Error(err)
		h.writeError(w, http.StatusBadRequest, "bad input JSON")
		return
	}

	if err := h.elastic.put(docs); err != nil {
		log.Error(err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
}

func (h *apiHandler) writeError(w http.ResponseWriter, code int, err string) {
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err)))
}

func (h *apiHandler) writeJSON(w http.ResponseWriter, conte int, object interface{}) error {
	w.Header().Add("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(object); err != nil {
		return wrapper.Wrap(err, "error while write json")
	}

	return nil
}

func (h *apiHandler) readKeys(private, public string) error {
	privateBytes, err := ioutil.ReadFile(private)
	if err != nil {
		return wrapper.Wrap(err, "error while read private key")
	}

	publicBytes, err := ioutil.ReadFile(public)
	if err != nil {
		return wrapper.Wrap(err, "error while read public key")
	}

	privateKey, err := parsePrivateKey(privateBytes)
	if err != nil {
		return wrapper.Wrap(err, "error while parse private key")
	}

	publicKey, err := parsePublicKey(publicBytes)
	if err != nil {
		return wrapper.Wrap(err, "error while parse public key")
	}

	h.signKey = privateKey
	h.verifyKey = publicKey

	return nil
}
