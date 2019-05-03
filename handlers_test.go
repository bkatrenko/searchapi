package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joho/godotenv"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTokenV1(t *testing.T) {
	Convey("Test get token", t, func() {
		handlers, err := newAPIHandler(testElasticAddr, tokenSignKey, tokenVerifyKey, tokenExpireAt)
		if err != nil {
			t.Fatal(err)
		}

		if err := handlers.elastic.deleteIndex(elasticIndexName); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("GET", "/v1/auth?creds=someid:somesecret", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handlers.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		var token Token
		if err := json.Unmarshal(rr.Body.Bytes(), &token); err != nil {
			t.Fatal(err)
		}

		if token.Token == "" {
			t.Fatal("error while get token: token is empty")
		}
	})
}

func TestPostProductV1(t *testing.T) {
	Convey("Test post product", t, func() {
		handlers, err := newAPIHandler(testElasticAddr, tokenSignKey, tokenVerifyKey, tokenExpireAt)
		if err != nil {
			t.Fatal(err)
		}

		if err := handlers.elastic.deleteIndex(elasticIndexName); err != nil {
			t.Fatal(err)
		}

		testDoc := Doc{
			Name:     "test_doc",
			Brand:    "test_brand",
			KeyWords: []string{"key", "word"},
		}

		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(testDoc)
		req, err := http.NewRequest("POST", "/v1/product", buf)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handlers, err = newAPIHandler(testElasticAddr, tokenSignKey, tokenVerifyKey, tokenExpireAt)
		if err != nil {
			t.Fatal(err)
		}
		handlers.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		shoes, err := handlers.elastic.get(queryParams{
			q:      testDoc.Name,
			limit:  1,
			offset: 0,
		})
		if err != nil {
			t.Fatal(err)
		}

		So(shoes[0].Name, ShouldEqual, testDoc.Name)
	})
}

func TestGetProductsV1(t *testing.T) {
	Convey("Test get products", t, func() {
		handlers, err := newAPIHandler(testElasticAddr, tokenSignKey, tokenVerifyKey, tokenExpireAt)
		if err != nil {
			t.Fatal(err)
		}

		if err := handlers.elastic.deleteIndex(elasticIndexName); err != nil {
			t.Fatal(err)
		}

		testDocs := []Doc{
			Doc{
				Name:     "aa",
				Brand:    "abc",
				KeyWords: []string{"summer", "cool"},
			},
			Doc{
				Name:     "bb",
				Brand:    "def",
				KeyWords: []string{"winter"},
			},

			Doc{
				Name:     "ght",
				Brand:    "yui",
				KeyWords: []string{"usa", "uk"},
			},
		}

		if err := handlers.elastic.put(testDocs); err != nil {
			t.Fatal(err)
		}

		returnedDocs := []Doc{}
		req, err := http.NewRequest("GET", "/v1/products?asc=false", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handlers.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if err := json.Unmarshal(rr.Body.Bytes(), &returnedDocs); err != nil {
			t.Fatal(err)
		}

		So(len(returnedDocs), ShouldEqual, 3)
	})
}

func TestMain(m *testing.M) {
	godotenv.Load(envFile)
}
