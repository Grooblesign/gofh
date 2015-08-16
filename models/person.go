package models

import (
	"strings"
)

type Person struct {
	Id        int64  `json:"id"`
	FatherId  int64  `json:"fatherId"`
	Forenames string `json:"firstnames"`
	Gender    string `json:"gender"`
	MotherId  int64  `json:"motherId"`
	Notes     string `json:"notes"`
	Surname   string `json:"surname"`
}
func (person *Person) GetFullName() string {
	return strings.TrimLeft(strings.TrimRight(person.Forenames + " " + person.Surname, " "), " ")
}

type ByName []Person

func (s ByName) Len() int {
	return len(s)
}
func (s ByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByName) Less(i, j int) bool {
	iName := s[i].GetFullName()
	jName := s[j].GetFullName()

	return (iName < jName)
}

