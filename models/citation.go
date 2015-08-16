package models

import (
)

type Citation struct {
	Id       int64  `json:"id"`
	Details  string `json:"details"`
	SourceId int64  `json:"sourceId"`
}


