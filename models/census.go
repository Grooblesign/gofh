package models

import (
)

type Census struct {
	Id    int64  `json:"id"`
	Date  string `json:"date"`
	Title string `json:"title"`
}
	
type CensusHousehold struct {
	Id         int64  `json:"id"`
	Address    string `json:"address"`
	CensusId   int64  `json:"censusId"`
	CitationId int64  `json:"citationId"`
	Folio      string `json:"folio"`
	Notes      string `json:"notes"`
	Page       string `json:"page"`
	Piece      string `json:"piece"`

	Persons []CensusHouseholdPerson `json:"persons"`
}

type CensusHouseholdPerson struct {
	Id                 int64  `json:"id"`
	Age                string `json:"age"`
	Birthplace         string `json:"birthplace"`
	CensusHouseholdId  int64  `json:"censusHouseholdId"`
	Name               string `json:"name"`
	Occupation         string `json:"occupation"`
	PersonId           int64  `json:"personId"`
	RelationshipToHead string `json:"relationshipToHead"`
	RowNumber          int64  `json:"rowNumber"`
	Status             string `json:"status"`
}
