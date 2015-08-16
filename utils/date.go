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

	// fmt.Println("-------------")
	// fmt.Println(date)

	var adjustment int64
	var err error
	var match bool
	var value int64 = 99999999
	var value1 int64
	var value2 int64

	months := "JANFEBMARAPRMAYJUNJULAUGSEPOCTNOVDEC"

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
		value1, err = strconv.ParseInt(date[0:4], 10, 64)
		value = value1 * 10000
		CheckErr(err)
	}

	match, _ = regexp.MatchString("^Q[0-9]{1} ([0-9]{4})$", date)
	if match {
		value1, err = strconv.ParseInt(date[3:7], 10, 64)
		CheckErr(err)
		value2, err = strconv.ParseInt(date[1:2], 10, 64)
		CheckErr(err)

		value = (value1 * 10000) + 100 + ((value2 - 1) * 300)
	}

	match, _ = regexp.MatchString("^[0-9]{2} [a-zA-Z]{3} ([0-9]{4})$", date)
	if match {
		value1, err = strconv.ParseInt(date[7:11], 10, 64)
		value = (value1 * 10000)
		value2, err = strconv.ParseInt(date[0:2], 10, 64)
		value = (value + value2)
		value = value + (int64)(((strings.Index(months, strings.ToUpper(date[3:6]))/3)+1)*100)
	}

	match, _ = regexp.MatchString("^[a-zA-Z]{3} ([0-9]{4})", date)
	if match {
		value1, err = strconv.ParseInt(date[4:8], 10, 64)
		value = (value1 * 10000)
		value = value + (int64)(((strings.Index(months, strings.ToUpper(date[0:3]))/3)+1)*100)
	}

	value = value + adjustment

	// fmt.Println(value)
	// fmt.Println("-------------")

	return value
}


