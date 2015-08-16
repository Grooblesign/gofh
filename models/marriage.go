package models

import (
)

type Marriage struct {
	Id         int64  `json:"id"`
	CitationId int64  `json:"citationId"`
	Date       string `json:"date"`
	HusbandId  int64  `json:"husbandId"`
	Location   string `json:"location"`
	Notes      string `json:"notes"`
	WifeId     int64  `json:"wifeId"`
}


