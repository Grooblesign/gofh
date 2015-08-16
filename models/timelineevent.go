package models

import (
	"../utils"
)

type TimelineEvent struct {
	Id        int64  `json:"id"`
	Date      string `json:"date"`
	Details   string `json:"details"`
	EventType string `json:"type"`
	IsPrimary bool   `json:"isPrimary"`
	Location  string `json:"location"`
}

type ByDate []TimelineEvent

func (s ByDate) Len() int {
	return len(s)
}
func (s ByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByDate) Less(i, j int) bool {
	iDateValue := utils.GetDateValue(s[i].Date)
	jDateValue := utils.GetDateValue(s[j].Date)

	return (iDateValue < jDateValue)
}
