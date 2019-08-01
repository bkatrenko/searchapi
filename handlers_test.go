package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testToken = "SEARCH_API_TEST_TOKEN"
)

var (
	// TestPut +++++++++++++++++++++++++++++++++++++++
	testPostDoc = []Doc{
		Doc{
			Name:     "test_doc",
			Brand:    "test_brand",
			KeyWords: []string{"key", "word"},
		},
	}

	checkPostQueryParams = getQueryParams{
		q:      testPostDoc[0].Name,
		limit:  10,
		offset: 0,
		sortBy: "name",
		order:  "desc",
	}
	// TestPut +++++++++++++++++++++++++++++++++++++++

	// TestGet +++++++++++++++++++++++++++++++++++++++
	testGetDocs = []Doc{
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
	// TestGet +++++++++++++++++++++++++++++++++++++++
)

func TestGetTokenV1(t *testing.T) {
	Convey("Test get token", t, func() {
		handlers := newTestHandlers(t)

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
		handlers := newTestHandlers(t)

		buf := new(bytes.Buffer)

		if err := json.NewEncoder(buf).Encode(testPostDoc); err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/v1/products", buf)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", os.Getenv(testToken))

		rr := httptest.NewRecorder()
		handlers, err = newAPIHandler(os.Getenv(elasticAddr), os.Getenv(tokenSignKey), os.Getenv(tokenVerifyKey), os.Getenv(tokenExpireAt))
		if err != nil {
			t.Fatal(err)
		}
		handlers.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		<-time.After(time.Second)
		shoes, err := handlers.elastic.get(&checkPostQueryParams)
		if err != nil {
			t.Fatal(err)
		}

		So(shoes.Hits[0].Name, ShouldEqual, testPostDoc[0].Name)
	})
}

func TestGetProductsV1(t *testing.T) {
	Convey("Test get products", t, func() {
		handlers := newTestHandlers(t)

		if err := handlers.elastic.put(testGetDocs, waitingUntilIndexed); err != nil {
			t.Fatal(err)
		}
		<-time.After(time.Second)
		returnedDocs := searchResults{}
		req, err := http.NewRequest("GET", "/v1/products", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", os.Getenv(testToken))

		rr := httptest.NewRecorder()
		handlers.router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		if err := json.Unmarshal(rr.Body.Bytes(), &returnedDocs); err != nil {
			t.Fatal(err)
		}

		So(returnedDocs.Total, ShouldEqual, 3)
		So(len(returnedDocs.Hits), ShouldEqual, 3)
	})
}

func newTestHandlers(t *testing.T) apiHandler {
	handlers, err := newAPIHandler(os.Getenv(elasticAddr), os.Getenv(tokenSignKey), os.Getenv(tokenVerifyKey), os.Getenv(tokenExpireAt))
	if err != nil {
		t.Fatal(err)
	}

	if err := handlers.elastic.deleteIndex(shoesIndex); err != nil {
		t.Fatal(err)
	}

	return handlers
}
