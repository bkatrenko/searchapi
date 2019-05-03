package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/olivere/elastic/v7"
	wrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	elasticIndexName = "documents"
)

type elasticEngine struct {
	client *elastic.Client
}

func newElasticEngine(elasticAddr string) (elasticEngine, error) {
	var (
		elasticClient *elastic.Client
		err           error
	)

	for i := 0; i < 10; i++ {
		elasticClient, err = elastic.NewClient(
			elastic.SetURL(elasticAddr),
			elastic.SetSniff(false),
		)
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	if elasticClient == nil {
		return elasticEngine{}, wrapper.Wrap(err, "can't connect to elastic")
	}

	client := elasticEngine{
		client: elasticClient,
	}

	if err := client.createIndex(); err != nil {
		return elasticEngine{}, wrapper.Wrap(err, "error while create index")
	}

	return client, nil
}

func (ee *elasticEngine) put(shoes []Doc) error {
	bulk := ee.client.
		Bulk().
		Index(elasticIndexName)

	for _, shoe := range shoes {
		shoe.ID = uuid.New().String()
		shoe.CreatedAt = time.Now().UTC()
		bulk.Add(elastic.NewBulkIndexRequest().Id(shoe.ID).Doc(shoe))
	}

	if _, err := bulk.Do(context.Background()); err != nil {
		return wrapper.Wrap(err, "error while insert value to elastic")
	}

	return nil
}

func (ee *elasticEngine) get(params queryParams) ([]Doc, error) {
	var esQuery elastic.Query

	log.Infof("start query with params: %+v", params)

	switch params.q {
	case "":
		esQuery = elastic.NewMatchAllQuery()
	default:
		esQuery = elastic.NewMultiMatchQuery(params.q, "name", "brand", "key_words").
			Fuzziness("2").
			MinimumShouldMatch("2")
	}

	esQuery = addFilter(esQuery, params.filterField, params.filterValue)

	result, err := ee.client.Search().
		Index(elasticIndexName).
		Query(esQuery).
		SortBy(elastic.NewFieldSort(params.sortBy).Order(params.asc).SortMode("min")).
		From(params.offset).Size(params.limit).
		Do(context.Background())
	if err != nil {
		return nil, wrapper.Wrap(err, "error while do shoes query")
	}

	res := []Doc{}
	for _, hit := range result.Hits.Hits {
		doc := Doc{}
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			return nil, wrapper.Wrap(err, "error while unmarshal elastic doc")
		}

		res = append(res, doc)
	}

	return res, nil
}

func (ee *elasticEngine) createIndex() error {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := ee.client.IndexExists(elasticIndexName).Do(context.Background())
	if err != nil {
		return wrapper.Wrap(err, "error while check index exists")
	}
	if !exists {
		// Create a new index.
		mapping := `
{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
			"properties":{
				"id": {
					"type":"text"
				},
				"created_at":{
					"type":"date"
				},
				"name":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"brand":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"key_words":{
					"type":"keyword"
				}
		}
	}
}
`

		_, err := ee.client.CreateIndex(elasticIndexName).Body(mapping).Do(context.Background())
		if err != nil {
			return wrapper.Wrap(err, "error while create index")
		}
	}

	return nil
}

func (ee *elasticEngine) deleteIndex(index string) error {
	resp, err := ee.client.DeleteIndex(index).
		Do(context.Background())
	if err != nil {
		return wrapper.Wrap(err, "error while delete index")
	}

	if !resp.Acknowledged {
		return errors.New("index not deleted: Acknowledged is false")
	}
	return nil
}

func addFilter(q elastic.Query, filter, with string) elastic.Query {
	switch filter {
	case "brand", "name", "key_words":
		return buildFilters(q, filter, with)
	default:
		return q
	}
}

func buildFilters(q elastic.Query, filter, with string) elastic.Query {
	return elastic.NewBoolQuery().
		Must(q).
		Filter(elastic.NewMatchQuery(filter, with))
}
