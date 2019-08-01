package main

import (
	"encoding/json"
	"time"
)

type envelopeResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     json.RawMessage `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

type createIndexResponse struct {
	Error struct {
		RootCause []struct {
			Type      string `json:"type"`
			Reason    string `json:"reason"`
			IndexUUID string `json:"index_uuid"`
			Index     string `json:"index"`
		} `json:"root_cause"`
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		IndexUUID string `json:"index_uuid"`
		Index     string `json:"index"`
	} `json:"error"`
	Status int `json:"status"`
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

// SearchResults wraps the Elasticsearch search response.
//
type searchResults struct {
	Total int    `json:"total"`
	Hits  []*Hit `json:"hits"`
}

// Hit wraps the document returned in search response.
//
type Hit struct {
	Score float64
	Doc
}

type Doc struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand"`
	KeyWords  []string  `json:"key_words"`
}

type Token struct {
	Token string `json:"token"`
}
