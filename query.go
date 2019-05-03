package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type queryParams struct {
	q           string
	sortBy      string
	asc         bool
	filterField string
	filterValue string
	limit       int
	offset      int
}

func (qp *queryParams) fill(r *http.Request) error {
	values := r.URL.Query()
	q, ok := values["q"]
	if ok {
		if len(q) >= 1 {
			qp.q = q[0]
		}
	}

	sortBy, ok := values["sort_by"]
	if ok {
		if len(sortBy) >= 1 {
			qp.sortBy = sortBy[0]
		}
	}

	asc, ok := values["asc"]
	if ok {
		if len(asc) > 0 {
			ascVal, err := strconv.ParseBool(asc[0])
			if err != nil {
				return errors.New("can't parse asc value")
			}
			qp.asc = ascVal
		}
	}

	filter, ok := values["filter"]
	if ok {
		if len(filter) > 0 {
			filterParams := strings.Split(filter[0], ":")
			if len(filterParams) < 2 {
				return errors.New("can't parse filter params, expect field:value")
			}
			qp.filterField = filterParams[0]
			qp.filterValue = filterParams[1]
		}
	}

	limit, ok := values["limit"]
	if ok {
		if len(limit) > 0 {
			limitVal, err := strconv.ParseInt(limit[0], 10, 64)
			if err != nil {
				return errors.New("error while parse limit")
			}
			qp.limit = int(limitVal)
		}
	}

	offset, ok := values["offset"]
	if ok {
		if len(offset) > 0 {
			offsetVal, err := strconv.ParseInt(offset[0], 10, 64)
			if err != nil {
				return errors.New("error while parse limit")
			}
			qp.offset = int(offsetVal)
		}
	}

	return nil
}

func (qp *queryParams) validate() error {
	if qp.sortBy == "" {
		qp.sortBy = "name"
	}

	if qp.limit <= 0 {
		qp.limit = 10
	}

	if qp.offset < 0 {
		qp.offset = 0
	}

	return nil
}
