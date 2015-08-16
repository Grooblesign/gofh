package models

import (
)

type Event struct {
	Id         int64  `json:"id"`
	CitationId int64  `json:"citationId"`
	Date       string `json:"date"`
	Details    string `json:"details"`
	EventType  string `json:"type"`
	IsPrimary  bool   `json:"isPrimary"`
	Location   string `json:"location"`
	Notes      string `json:"notes"`
	PersonId   int64  `json:"personId"`
}


