package main

import (
	"time"
)

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
