package main

import (
	"os"
	"testing"
	"time"

	"github.com/icrowley/fake"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	waitingUntilIndexed = true
	notWaiting          = false
)

var (
	// TestPut +++++++++++++++++++++++++++++++++++++++
	testPutDocs = []Doc{
		Doc{
			Name:  "shoe",
			Brand: "bkatrenko",
		},
	}

	checkPutQueryParams = getQueryParams{
		sortBy: "name",
		order:  "desc",
		limit:  1,
	}
	// TestPut +++++++++++++++++++++++++++++++++++++++

	// TestSort ++++++++++++++++++++++++++++++++++++++
	testSortDocs = []Doc{
		Doc{
			Name:  "a test sort",
			Brand: "bkatrenko",
		},

		Doc{
			Name:  "b test sort",
			Brand: "not me",
		},
	}

	checkSortQueryParams = getQueryParams{
		q:      "test sort",
		sortBy: "name",
		order:  "desc",
		limit:  2,
	}
	// TestSort ++++++++++++++++++++++++++++++++++++++

	// TestFilter ++++++++++++++++++++++++++++++++++++
	testFilterDocs = []Doc{
		Doc{
			Name:  "a test filter",
			Brand: "bkatrenko",
		},

		Doc{
			Name:  "b test filter",
			Brand: "not me",
		},
	}

	checkFilterQueryParams = getQueryParams{
		q:           "test filter",
		sortBy:      "name",
		order:       "asc",
		filterField: "brand",
		filterValue: "bkatrenko",
		limit:       2,
	}
)

func TestPut(t *testing.T) {
	Convey("check elastic put", t, func() {
		engine := newTestEngine(t)

		if err := engine.put(testPutDocs, waitingUntilIndexed); err != nil {
			t.Fatal(err)
		}

		shoes, err := engine.get(&checkPutQueryParams)
		if err != nil {
			t.Fatal(err)
		}

		So(len(shoes.Hits), ShouldBeGreaterThan, 0)
		So(shoes.Hits[0].Name, ShouldEqual, "shoe")
	})
}

func TestSort(t *testing.T) {
	Convey("Check elastic sort", t, func() {
		engine := newTestEngine(t)

		if err := engine.put(testSortDocs, waitingUntilIndexed); err != nil {
			t.Fatal(err)
		}

		shoes, err := engine.get(&checkSortQueryParams)
		if err != nil {
			t.Fatal(err)
		}

		So(shoes.Hits[0].Name, ShouldEqual, "a test sort")
		So(shoes.Hits[1].Name, ShouldEqual, "b test sort")

		checkSortQueryParams.order = "asc"
		shoes, err = engine.get(&checkSortQueryParams)
		if err != nil {
			t.Fatal(err)
		}

		So(shoes.Hits[0].Name, ShouldEqual, "b test sort")
		So(shoes.Hits[1].Name, ShouldEqual, "a test sort")
	})
}

func TestFilter(t *testing.T) {
	Convey("Check elastic filter", t, func() {
		engine := newTestEngine(t)

		if err := engine.put(testFilterDocs, waitingUntilIndexed); err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Second)

		shoes, err := engine.get(&checkFilterQueryParams)
		if err != nil {
			t.Fatal(err)
		}

		So(len(shoes.Hits), ShouldEqual, 1)
	})
}

func BenchmarkPut(b *testing.B) {
	for i := 0; i < b.N; i++ {
		docs := []Doc{}
		for i := 0; i < 500; i++ {
			docs = append(docs, Doc{
				Name:     fake.Product(),
				Brand:    fake.Company(),
				KeyWords: []string{fake.City(), fake.Country(), fake.DomainName()},
			})
		}

		engine, err := newElasticEngine(os.Getenv(elasticAddr))
		if err != nil {
			b.Fatal(err)
		}

		if err := engine.put(docs, notWaiting); err != nil {
			b.Fatal(err)
		}
	}
}

func newTestEngine(t *testing.T) searchEngine {
	engine, err := newElasticEngine(os.Getenv(elasticAddr))
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.deleteIndex(shoesIndex); err != nil {
		t.Fatal(err)
	}

	if err := engine.createIndex(shoesIndex, shoesMapping); err != nil {
		t.Fatal(err)
	}

	return engine
}
