package utils

import (
	"regexp"
	"strconv"
	"strings"
)

//-----------------------------------------------------------------------------
// GetDateValue
//-----------------------------------------------------------------------------

func GetDateValue(date string) int64 {

	var adjustment int64
	var dateValue int64 = 99999999
	var err error
	var match bool
	var value int64

	const months = "JANFEBMARAPRMAYJUNJULAUGSEPOCTNOVDEC"

	dashPos := strings.Index(date, "-")
	if dashPos > -1 {
		date = strings.TrimRight(date[:dashPos], " ")
	}

	match, _ = regexp.MatchString("^bef", date)
	if match {
		adjustment = -1
		date = date[4:]
	}

	match, _ = regexp.MatchString("^aft", date)
	if match {
		adjustment = 1
		date = date[4:]
	}

	match, _ = regexp.MatchString("^[0-9]{4}", date)
	if match {
		value, err = strconv.ParseInt(date[0:4], 10, 64)
		dateValue = value * 10000
		CheckErr(err)
	}

	match, _ = regexp.MatchString("^Q[0-9]{1} ([0-9]{4})$", date)
	if match {
		value, err = strconv.ParseInt(date[3:7], 10, 64)
		CheckErr(err)
		dateValue = (value * 10000) + 100

		value, err = strconv.ParseInt(date[1:2], 10, 64)
		CheckErr(err)
		dateValue = dateValue + ((value - 1) * 300)
	}

	match, _ = regexp.MatchString("^[0-9]{2} [a-zA-Z]{3} ([0-9]{4})$", date)
	if match {
		value, err = strconv.ParseInt(date[7:11], 10, 64)
		dateValue = (value * 10000)
		value, err = strconv.ParseInt(date[0:2], 10, 64)
		dateValue = (dateValue + value)
		dateValue = dateValue + (int64)(((strings.Index(months, strings.ToUpper(date[3:6]))/3)+1)*100)
	}

	match, _ = regexp.MatchString("^[a-zA-Z]{3} ([0-9]{4})", date)
	if match {
		value, err = strconv.ParseInt(date[4:8], 10, 64)
		dateValue = (value * 10000)
		dateValue = dateValue + (int64)(((strings.Index(months, strings.ToUpper(date[0:3]))/3)+1)*100)
	}

	dateValue = dateValue + adjustment

	return dateValue
}
