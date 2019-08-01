package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	wrapper "github.com/pkg/errors"
)

const (
	matchAllTemplate = `"match_all":{}`

	multiMatchTemplate = `
						"multi_match" : {
							"query":    "%s", 
							"fields": ["name", "brand", "key_words"] 
				  		}`

	filterTemplate = `
						{
							"term": {
								"%s": "%s"
							}
						}`

	shoesSearch = `
						{
							"query": {
							"bool": {
							"must": [
								{
									%s
								}
							],
							"filter": [%s] 
							}
						},
							"sort" : { 
								"%s" : "%s"
							},
							"size" : %d,
							"from": %d
						}`
)

type getQueryParams struct {
	q           string
	sortBy      string
	order       string
	filterField string
	filterValue string
	limit       int
	offset      int
}

type putQueryParams struct {
	waitingForIndexed bool
}

func (pqp *putQueryParams) fill(r *http.Request) error {
	values := r.URL.Query()
	q, _ := values["waiting"]
	if len(q) == 0 {
		pqp.waitingForIndexed = false
		return nil
	}

	waitingValue, err := strconv.ParseBool(q[0])
	if err != nil {
		return wrapper.Wrap(err, "error while parse 'waiting for' value")
	}

	pqp.waitingForIndexed = waitingValue
	return nil
}

func (qp *getQueryParams) fill(r *http.Request) error {
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

	order, ok := values["order"]
	if ok {
		qp.order = order[0]
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

func (qp *getQueryParams) validateAndFill() error {
	switch qp.sortBy {
	case "":
		qp.sortBy = "created_at"
	case "name":
	case "created_at":
	case "brand":
	default:
		return errors.New("bad sort_by param")
	}

	switch qp.filterField {
	case "":
	case "name":
	case "brand":
	case "created_at":
	default:
		return errors.New("bad filter field value")
	}

	switch qp.order {
	case "":
		qp.order = "desc"
	case "asc":
	case "desc":
	default:
		return errors.New("bad order value")
	}

	if qp.limit <= 0 {
		qp.limit = 25
	}

	if qp.offset < 0 {
		qp.offset = 0
	}

	return nil
}

func (qp *getQueryParams) build() io.Reader {
	var b strings.Builder

	var searchQuery string
	var filters string

	switch qp.q {
	case "":
		searchQuery = matchAllTemplate
	default:
		searchQuery = fmt.Sprintf(multiMatchTemplate, qp.q)
	}

	switch qp.filterField {
	case "":
		filters = ""
	default:
		filters = fmt.Sprintf(filterTemplate, qp.filterField, qp.filterValue)
	}

	b.WriteString(fmt.Sprintf(shoesSearch, searchQuery, filters, qp.sortBy, qp.order, qp.limit, qp.offset))
	return strings.NewReader(b.String())
}
