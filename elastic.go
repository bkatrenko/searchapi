package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	wrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	shoesIndex = "shoes_mapping"

	shoesMapping = `
	{
		"settings":{
			"number_of_shards":5,
			"number_of_replicas":0,
			"analysis":{
				"filter":{
				   "stemmer":{
					  "type":"stemmer",
					  "language":"english"
				   },
				   "stopwords":{
					  "type":"stop",
					  "stopwords":[
						 "_english_"
					  ]
				   }
				},
				"analyzer":{
				   "shoe_analyzer":{
					  "filter":[
						 "stopwords",
						 "lowercase",
						 "stemmer"
					  ],
					  "type":"custom",
					  "tokenizer":"standard"
				   }
				}
			 }
		},
		"mappings":{
			    "dynamic" : false,
				"properties":{
					"id": {
						"type":"text"
					},
					"created_at":{
						"type":"date"
					},
					"name":{
						"type":"text",
						"fielddata": true,
						"analyzer":"shoe_analyzer",
						"search_analyzer":"shoe_analyzer"
					},
					"brand":{
						"type":"text",
						"fielddata": true,
						"analyzer":"shoe_analyzer",
						"search_analyzer":"shoe_analyzer"
					},
					"key_words":{
						"type":"keyword"
					}
			}
		}
	}`
)

type searchEngine interface {
	createIndex(index, mapping string) error
	deleteIndex(index string) error

	put(shoes []Doc, waitingUntilIndexed bool) error
	get(params *getQueryParams) (searchResults, error)
}

type elasticEngine struct {
	client *elasticsearch.Client
}

func newElasticEngine(elasticAddr string) (searchEngine, error) {
	esClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, wrapper.Wrap(err, "error creating the client")
	}

	client := elasticEngine{
		client: esClient,
	}

	if err := client.createIndex(shoesIndex, shoesMapping); err != nil {
		return nil, wrapper.Wrap(err, "error while create index")
	}

	return &client, nil
}

func (ee *elasticEngine) put(shoes []Doc, waitingUntilIndexed bool) error {
	var buf bytes.Buffer
	var blk *bulkResponse

	raw := map[string]interface{}{}

	for _, a := range shoes {
		a.CreatedAt = time.Now()
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" }}%s`, shoesIndex, "\n"))
		data, err := json.Marshal(a)
		if err != nil {
			return wrapper.Wrap(err, "error while encode doc")
		}

		data = append(data, "\n"...)

		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}

	args := []func(*esapi.BulkRequest){ee.client.Bulk.WithIndex(shoesIndex)}
	if waitingUntilIndexed {
		args = append(args, ee.client.Bulk.WithRefresh("wait_for"))
	}

	res, err := ee.client.Bulk(bytes.NewReader(buf.Bytes()), args...)
	if err != nil {
		return wrapper.Wrap(err, "error while indexind butch")
	}

	if res.IsError() {
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			return wrapper.Wrap(err, "failure to to parse response body")
		}

		return errors.New(fmt.Sprintf("error: [%d] %s: %s",
			res.StatusCode,
			raw["error"].(map[string]interface{})["type"],
			raw["error"].(map[string]interface{})["reason"]))

	}

	if err := json.NewDecoder(res.Body).Decode(&blk); err != nil {
		return wrapper.Wrap(err, "failure to to parse response body")
	}

	for _, d := range blk.Items {
		if d.Index.Status > 201 {
			log.Errorf("error: [%d]: %s: %s: %s: %s",
				d.Index.Status,
				d.Index.Error.Type,
				d.Index.Error.Reason,
				d.Index.Error.Cause.Type,
				d.Index.Error.Cause.Reason,
			)
		}
	}

	log.Infof("bulk write complete for %d items", len(shoes))
	return nil
}

func (ee *elasticEngine) get(params *getQueryParams) (searchResults, error) {
	var results searchResults

	res, err := ee.client.Search(
		ee.client.Search.WithIndex(shoesIndex),
		ee.client.Search.WithBody(params.build()),
	)
	if err != nil {
		return searchResults{}, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return searchResults{}, wrapper.Wrapf(err, "can't unmarshal returned error: %v, %s", err, res.String())
		}
		return searchResults{}, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return searchResults{}, err
	}

	results.Total = r.Hits.Total.Value

	if len(r.Hits.Hits) == 0 {
		results.Hits = []*Hit{}
		return searchResults{}, nil
	}

	for _, hit := range r.Hits.Hits {
		var h Hit
		h.ID = hit.ID

		if err := json.Unmarshal(hit.Source, &h); err != nil {
			return searchResults{}, err
		}

		results.Hits = append(results.Hits, &h)
	}

	return results, nil
}

func (ee *elasticEngine) createIndex(index, mapping string) error {
	res, err := ee.client.Indices.Create(index, ee.client.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return wrapper.Wrap(err, "error while create index")
	}

	if res.IsError() {
		createIndexResponse := createIndexResponse{}

		if err := json.NewDecoder(res.Body).Decode(&createIndexResponse); err != nil {
			log.Error(err, string(res.String()))
			return errors.New(res.String())
		}

		if len(createIndexResponse.Error.RootCause) < 0 {
			return errors.New(res.String())
		}

		if createIndexResponse.Error.RootCause[0].Type == "resource_already_exists_exception" {
			return nil
		}

		return errors.New(res.String())
	}

	return nil
}

func (ee *elasticEngine) deleteIndex(index string) error {
	res, err := ee.client.Indices.Delete([]string{index})
	if err != nil {
		return wrapper.Wrap(err, "error while create index")
	}

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
}
