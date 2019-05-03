package main

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testElasticAddr = "http://localhost:9200"
)

func TestPut(t *testing.T) {
	Convey("Check elastic put", t, func() {
		engine, err := newElasticEngine(testElasticAddr)
		if err != nil {
			t.Fatal(err)
		}

		if err := engine.put([]Doc{
			Doc{
				Name:  "shoe",
				Brand: "bkatrenko",
			},
		}); err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Second)

		//shoes, err := engine.get("shoe", "name", true, "", "", 1, 0)
		shoes, err := engine.get(queryParams{
			q:      "shoe",
			sortBy: "name",
			asc:    true,
			limit:  1,
			offset: 0,
		})
		if err != nil {
			t.Fatal(err)
		}

		So(len(shoes), ShouldBeGreaterThan, 0)
		So(shoes[0].Name, ShouldEqual, "shoe")
	})
}

func TestSort(t *testing.T) {
	Convey("Check elastic sort", t, func() {
		engine, err := newElasticEngine(testElasticAddr)
		if err != nil {
			t.Fatal(err)
		}

		if err := engine.deleteIndex(elasticIndexName); err != nil {
			t.Fatal(err)
		}

		if err := engine.createIndex(); err != nil {
			t.Fatal(err)
		}

		if err := engine.put([]Doc{
			Doc{
				Name:  "a test sort",
				Brand: "bkatrenko",
			},

			Doc{
				Name:  "b test sort",
				Brand: "not me",
			},
		}); err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Second)

		shoes, err := engine.get(queryParams{
			q:      "test sort",
			sortBy: "name",
			asc:    true,
			limit:  2,
			offset: 0,
		})
		if err != nil {
			t.Fatal(err)
		}

		So(shoes[0].Name, ShouldEqual, "a test sort")
		So(shoes[1].Name, ShouldEqual, "b test sort")

		shoes, err = engine.get(queryParams{
			q:      "test sort",
			sortBy: "name",
			asc:    false,
			limit:  2,
			offset: 0,
		})
		if err != nil {
			t.Fatal(err)
		}

		So(shoes[0].Name, ShouldEqual, "b test sort")
		So(shoes[1].Name, ShouldEqual, "a test sort")
	})
}

func TestFilter(t *testing.T) {
	Convey("Check elastic filter", t, func() {
		engine, err := newElasticEngine(testElasticAddr)
		if err != nil {
			t.Fatal(err)
		}

		if err := engine.deleteIndex(elasticIndexName); err != nil {
			t.Fatal(err)
		}

		if err := engine.createIndex(); err != nil {
			t.Fatal(err)
		}

		if err := engine.put([]Doc{
			Doc{
				Name:  "a test filter",
				Brand: "bkatrenko",
			},

			Doc{
				Name:  "b test filter",
				Brand: "not me",
			},
		}); err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Second)

		shoes, err := engine.get(queryParams{
			q:           "test filter",
			sortBy:      "name",
			asc:         true,
			filterField: "brand",
			filterValue: "bkatrenko",
			limit:       2,
			offset:      0,
		})
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(shoes)
		So(len(shoes), ShouldEqual, 1)
	})
}
