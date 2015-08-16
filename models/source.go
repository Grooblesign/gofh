package models

import (
)

type Source struct {
	Id         int64  `json:"id"`
	Author     string `json:"author"`
	Date       string `json:"date"`
	Notes      string `json:"notes"`
	Publisher  string `json:"publisher"`
	SourceType string `json:"sourceType"`
	Title      string `json:"title"`
	Url        string `json:"url"`
}
