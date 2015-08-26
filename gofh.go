package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"./models"
	"./utils"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "Tazzle11!"
	DB_NAME     = "FamilyHistory"

	SERVER_PORT = 8880
)

//-----------------------------------------------------------------------------
// main
//-----------------------------------------------------------------------------

func main() {
	router := mux.NewRouter().StrictSlash(true)

	addRoutes(router)

	fmt.Println(fmt.Sprintf("Listening on %d...", SERVER_PORT))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", SERVER_PORT), router))
}

//-----------------------------------------------------------------------------
// addRoutes
//-----------------------------------------------------------------------------

func addRoutes(router *mux.Router) {

	// HTML views
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/view/censushousehold/{censusHouseholdId}", viewCensusHousehold).Methods("GET")
	router.HandleFunc("/view/person/{personId}", viewPerson).Methods("GET")
	router.HandleFunc("/view/person/{personId}/ancestors", viewPersonAncestors).Methods("GET")
	router.HandleFunc("/view/persons/{surname}", viewPersons).Methods("GET")

	// REST api
	router.HandleFunc("/census", censusIndex).Methods("GET")
	router.HandleFunc("/censushousehold", censusHouseholdIndex).Methods("GET")
	router.HandleFunc("/censushousehold", censusHouseholdCreate).Methods("POST")
	router.HandleFunc("/censushousehold", censusHouseholdUpdate).Methods("PUT")
	router.HandleFunc("/censushouseholdperson", censusHouseholdPersonIndex).Methods("GET")
	router.HandleFunc("/censushouseholdperson", censusHouseholdPersonCreate).Methods("POST")
	router.HandleFunc("/censushouseholdperson", censusHouseholdPersonUpdate).Methods("PUT")
	router.HandleFunc("/citation", citationIndex).Methods("GET")
	router.HandleFunc("/citation", citationCreate).Methods("POST")
	router.HandleFunc("/citation", citationUpdate).Methods("PUT")
	router.HandleFunc("/event/{eventId}", eventDelete).Methods("DELETE")
	router.HandleFunc("/event", eventCreate).Methods("POST")
	router.HandleFunc("/event", eventUpdate).Methods("PUT")
	router.HandleFunc("/marriage", marriageCreate).Methods("POST")
	router.HandleFunc("/marriage", marriageUpdate).Methods("PUT")
	router.HandleFunc("/person", personIndex).Methods("GET")
	router.HandleFunc("/person/{personId}", personById).Methods("GET")
	router.HandleFunc("/person/{personId}/censushousehold", personCensusHouseholds).Methods("GET")
	router.HandleFunc("/person/{personId}/child", personChildren).Methods("GET")
	router.HandleFunc("/person/{personId}/event", personEvents).Methods("GET")
	router.HandleFunc("/person/{personId}/marriage", personMarriage).Methods("GET")
	router.HandleFunc("/person", personCreate).Methods("POST")
	router.HandleFunc("/person", personUpdate).Methods("PUT")
	router.HandleFunc("/source", sourceIndex).Methods("GET")
	router.HandleFunc("/source", sourceCreate).Methods("POST")
	router.HandleFunc("/source", sourceUpdate).Methods("PUT")
}

//-----------------------------------------------------------------------------
// censusIndex
//-----------------------------------------------------------------------------

func censusIndex(w http.ResponseWriter, r *http.Request) {

	censuses := getCensuses()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(censuses)
}

//-----------------------------------------------------------------------------
// censusHouseholdIndex
//-----------------------------------------------------------------------------

func censusHouseholdIndex(w http.ResponseWriter, r *http.Request) {

	censusHouseholds := getCensusHouseholds(0)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(censusHouseholds)
}

//-----------------------------------------------------------------------------
// censusHouseholdCreate
//-----------------------------------------------------------------------------

func censusHouseholdCreate(w http.ResponseWriter, r *http.Request) {

	var censusHousehold models.CensusHousehold

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &censusHousehold); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()

	_, err = db.Exec("INSERT INTO censushousehold (address, censusid, citationid, folio, notes, page, piece) VALUES ($1, $2, $3, $4, $5, $6, $7)", censusHousehold.Address, censusHousehold.CensusId, censusHousehold.CitationId, censusHousehold.Folio, censusHousehold.Notes, censusHousehold.Page, censusHousehold.Piece)
	utils.CheckErr(err)

	db.Close()

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// censusHouseholdUpdate
//-----------------------------------------------------------------------------

func censusHouseholdUpdate(w http.ResponseWriter, r *http.Request) {

	var censusHousehold models.CensusHousehold

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &censusHousehold); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()

	_, err = db.Exec("UPDATE censushousehold SET address=$2, censusid=$3, citationid=$4, folio=$5, notes=$6, page=$7, piece=$8 WHERE id=$1", censusHousehold.Id, censusHousehold.Address, censusHousehold.CensusId, censusHousehold.CitationId, censusHousehold.Folio, censusHousehold.Notes, censusHousehold.Page, censusHousehold.Piece)
	utils.CheckErr(err)

	db.Close()
}

//-----------------------------------------------------------------------------
// censusHouseholdPersonIndex
//-----------------------------------------------------------------------------

func censusHouseholdPersonIndex(w http.ResponseWriter, r *http.Request) {

	db := openDB()

	rows, err := db.Query("SELECT id, age, birthplace, censushouseholdid, name, occupation, personid, relationshiptohead, rownumber, status FROM censushouseholdperson ORDER BY rownumber, id")
	utils.CheckErr(err)

	censusHouseholdPersons := []models.CensusHouseholdPerson{}

	for rows.Next() {
		var id sql.NullInt64
		var age sql.NullString
		var birthplace sql.NullString
		var censusHouseholdId sql.NullInt64
		var name sql.NullString
		var occupation sql.NullString
		var personid sql.NullInt64
		var relationshipToHead sql.NullString
		var rowNumber sql.NullInt64
		var status sql.NullString

		err = rows.Scan(&id, &age, &birthplace, &censusHouseholdId, &name, &occupation, &personid, &relationshipToHead, &rowNumber, &status)
		utils.CheckErr(err)

		censusHouseholdPerson := models.CensusHouseholdPerson{getInt64(id), getString(age), getString(birthplace), getInt64(censusHouseholdId), getString(name), getString(occupation), getInt64(personid), getString(relationshipToHead), getInt64(rowNumber), getString(status)}

		censusHouseholdPersons = append(censusHouseholdPersons, censusHouseholdPerson)
	}

	rows.Close()
	db.Close()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(censusHouseholdPersons)
}

//-----------------------------------------------------------------------------
// censusHouseholdPersonCreate
//-----------------------------------------------------------------------------

func censusHouseholdPersonCreate(w http.ResponseWriter, r *http.Request) {

	var censusHouseholdPerson models.CensusHouseholdPerson

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &censusHouseholdPerson); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()

	_, err = db.Exec("INSERT INTO censushouseholdperson (age, birthplace, censushouseholdid, name, occupation, personid, relationshiptohead, rownumber, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", censusHouseholdPerson.Age, censusHouseholdPerson.Birthplace, censusHouseholdPerson.CensusHouseholdId, censusHouseholdPerson.Name, censusHouseholdPerson.Occupation, censusHouseholdPerson.PersonId, censusHouseholdPerson.RelationshipToHead, censusHouseholdPerson.RowNumber, censusHouseholdPerson.Status)

	utils.CheckErr(err)

	db.Close()

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// censusHouseholdPersonUpdate
//-----------------------------------------------------------------------------

func censusHouseholdPersonUpdate(w http.ResponseWriter, r *http.Request) {

	var censusHouseholdPerson models.CensusHouseholdPerson

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &censusHouseholdPerson); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()

	_, err = db.Exec("UPDATE censushouseholdperson SET age=$2, birthplace=$3, censushouseholdid=$4, name=$5, occupation=$6, personid=$7, relationshiptohead=$8, rownumber=$9, status=$10 WHERE id=$1", censusHouseholdPerson.Id, censusHouseholdPerson.Age, censusHouseholdPerson.Birthplace, censusHouseholdPerson.CensusHouseholdId, censusHouseholdPerson.Name, censusHouseholdPerson.Occupation, censusHouseholdPerson.PersonId, censusHouseholdPerson.RelationshipToHead, censusHouseholdPerson.RowNumber, censusHouseholdPerson.Status)

	utils.CheckErr(err)

	db.Close()
}

//-----------------------------------------------------------------------------
// citationCreate
//-----------------------------------------------------------------------------

func citationCreate(w http.ResponseWriter, r *http.Request) {

	var citation models.Citation

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &citation); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()

	_, err = db.Exec("INSERT INTO citation (details, sourceid) VALUES ($1, $2)", citation.Details, citation.SourceId)

	utils.CheckErr(err)

	db.Close()

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// citationUpdate
//-----------------------------------------------------------------------------

func citationUpdate(w http.ResponseWriter, r *http.Request) {

	var citation models.Citation

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &citation); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("UPDATE citation SEt details=$1, sourceid=$2 WHERE id=$3", citation.Details, citation.SourceId, citation.Id)

	utils.CheckErr(err)
}

//-----------------------------------------------------------------------------
// citationIndex
//-----------------------------------------------------------------------------

func citationIndex(w http.ResponseWriter, r *http.Request) {

	db := openDB()
	defer db.Close()

	rows, err := db.Query("SELECT id, details, sourceid FROM citation ORDER BY id")
	utils.CheckErr(err)
	defer rows.Close()

	citations := []models.Citation{}

	for rows.Next() {
		var id sql.NullInt64
		var details sql.NullString
		var sourceId sql.NullInt64

		err = rows.Scan(&id, &details, &sourceId)
		utils.CheckErr(err)

		citation := models.Citation{getInt64(id), getString(details), getInt64(sourceId)}

		citations = append(citations, citation)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(citations)
}

//-----------------------------------------------------------------------------
// eventCreate
//-----------------------------------------------------------------------------

func eventCreate(w http.ResponseWriter, r *http.Request) {

	var event models.Event

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &event); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("INSERT INTO event (citationid, date, details, type, isprimary, location, notes, personid) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		event.CitationId, event.Date, event.Details, event.EventType, event.IsPrimary, event.Location, event.Notes, event.PersonId)

	utils.CheckErr(err)

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// eventDelete
//-----------------------------------------------------------------------------

func eventDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventId, err := strconv.ParseInt(vars["eventId"], 10, 64)
	utils.CheckErr(err)

	db := openDB()
	defer db.Close()

	_, err = db.Exec("DELETE FROM event WHERE id=$1", eventId)

	utils.CheckErr(err)
}

//-----------------------------------------------------------------------------
// eventUpdate
//-----------------------------------------------------------------------------

func eventUpdate(w http.ResponseWriter, r *http.Request) {

	var event models.Event

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &event); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("UPDATE event SET citationid=$2, date=$3, details=$4, type=$5, isprimary=$6, location=$7, notes=$8, personid=$9 WHERE id=$1",
		event.Id, event.CitationId, event.Date, event.Details, event.EventType, event.IsPrimary, event.Location, event.Notes, event.PersonId)

	utils.CheckErr(err)
}

//-----------------------------------------------------------------------------
// index
//-----------------------------------------------------------------------------

func index(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte("<html>"))
	w.Write([]byte("<head>"))

	writeStylesheet(w)

	w.Write([]byte("<H1>Go Family History!</H1>"))
	
	w.Write([]byte("<h2>Show surnames starting with:</h2>"))

	w.Write([]byte("<table>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<td>&nbsp;</td><td>A</td><td>B</td><td>C</td><td>D</td><td>E</td><td>F</td><td>G</td><td>H</td><td>I</td><td>J</td><td>K</td><td>L</td><td>M</td><td>N</td><td>O</td><td>P</td><td>Q</td><td>R</td><td>S</td><td>T</td><td>U</td><td>V</td><td>W</td><td>X</td><td>Y</td><td>Z</td>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("</table>"))

	w.Write([]byte("<H2>Top 30 surnames (total individuals):</H2>"))

	db := openDB()

	rows, err := db.Query("SELECT surname, COUNT(*) AS total FROM person GROUP BY surname ORDER BY total DESC")
	utils.CheckErr(err)

	w.Write([]byte("<table>"))

	count := 0
	for rows.Next() {
		count++
		if count <= 30 {
			if count-1 % 3 == 0 {
				w.Write([]byte("<tr>"))
			}

			var surname sql.NullString
			var total sql.NullInt64

			err = rows.Scan(&surname, &total)
			utils.CheckErr(err)

			w.Write([]byte("<td>"))
			w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/persons/%s'>%s</a>", getString(surname), getString(surname))))
			w.Write([]byte("</td>"))

			w.Write([]byte("<td>"))
			w.Write([]byte(fmt.Sprintf("%d", getInt64(total))))
			w.Write([]byte("</td>"))

			if count % 3 == 0 {
				w.Write([]byte("</tr>"))
			}
		} else {
			break
		}
	}

	w.Write([]byte("</table>"))

	w.Write([]byte("</head>"))
	w.Write([]byte("<body>"))

	rows.Close()
	db.Close()
}

//-----------------------------------------------------------------------------
// marriageCreate
//-----------------------------------------------------------------------------

func marriageCreate(w http.ResponseWriter, r *http.Request) {

	var marriage models.Marriage

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &marriage); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("INSERT INTO marriage (citationid, date, husbandid, location, notes, wifeid) VALUES ($1, $2, $3, $4, $5, $6)",
		marriage.CitationId, marriage.Date, marriage.HusbandId, marriage.Location, marriage.Notes, marriage.WifeId)

	utils.CheckErr(err)

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// marriageUpdate
//-----------------------------------------------------------------------------

func marriageUpdate(w http.ResponseWriter, r *http.Request) {

	var marriage models.Marriage

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &marriage); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("UPDATE marriage SET citationid=$2, date=$3, husbandid=$4, location=$5, notes=$6, wifeid=$7 WHERE id=$1",
		marriage.Id, marriage.CitationId, marriage.Date, marriage.HusbandId, marriage.Location, marriage.Notes, marriage.WifeId)

	utils.CheckErr(err)
}

//-----------------------------------------------------------------------------
// personIndex
//-----------------------------------------------------------------------------

func personIndex(w http.ResponseWriter, r *http.Request) {

	db := openDB()
	defer db.Close()

	surname := r.URL.Query().Get("surname")

	var rows *sql.Rows
	var err error

	if len(surname) == 0 {
		rows, err = db.Query("SELECT id, fatherid, forenames, gender, motherid, notes, surname FROM person ORDER BY id")
	} else {
		rows, err = db.Query("SELECT id, fatherid, forenames, gender, motherid, notes, surname FROM person WHERE surname = $1 ORDER BY id", surname)
	}

	utils.CheckErr(err)
	defer rows.Close()

	persons := []models.	Person{}

	for rows.Next() {
		persons = append(persons, getPersonFromRow(rows))
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(persons)
}

//-----------------------------------------------------------------------------
// personCensusHouseholds
//-----------------------------------------------------------------------------

func personCensusHouseholds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	censusHouseholds := getCensusHouseholdsForPerson(personId)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(censusHouseholds)
}

//-----------------------------------------------------------------------------
// personChildren
//-----------------------------------------------------------------------------

func personChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	persons := getChildrenForPerson(personId)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(persons)
}

//-----------------------------------------------------------------------------
// personCreate
//-----------------------------------------------------------------------------

func personCreate(w http.ResponseWriter, r *http.Request) {

	var person models.Person

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &person); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("INSERT INTO person (fatherid, forenames, gender, motherid, notes, surname) VALUES ($1, $2, $3, $4, $5, $6)",
		person.FatherId, person.Forenames, person.Gender, person.MotherId, person.Notes, person.Surname)

	utils.CheckErr(err)

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// personUpdate
//-----------------------------------------------------------------------------

func personUpdate(w http.ResponseWriter, r *http.Request) {

	var person models.Person

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &person); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("UPDATE person SET fatherid=$1, forenames=$2, gender=$3, motherid=$4, notes=$5, surname=$6 WHERE id=$7",
		person.FatherId, person.Forenames, person.Gender, person.MotherId, person.Notes, person.Surname, person.Id)

	utils.CheckErr(err)
}

//-----------------------------------------------------------------------------
// personEvents
//-----------------------------------------------------------------------------

func personEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	events := getEventsForPerson(personId)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(events)
}

//-----------------------------------------------------------------------------
// personMarriage
//-----------------------------------------------------------------------------

func personMarriage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	marriages := getMarriagesForPerson(personId)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(marriages)
}

//-----------------------------------------------------------------------------
// personById
//-----------------------------------------------------------------------------

func personById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	person := getPersonById(personId)

	if person != nil {
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(person)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

//-----------------------------------------------------------------------------
// sourceIndex
//-----------------------------------------------------------------------------

func sourceIndex(w http.ResponseWriter, r *http.Request) {

	db := openDB()
	defer db.Close()

	rows, err := db.Query("SELECT id, author, date, notes, publisher, type, title, url FROM source ORDER BY id")
	utils.CheckErr(err)
	defer rows.Close()

	sources := []models.Source{}

	for rows.Next() {
		var id sql.NullInt64
		var author sql.NullString
		var date sql.NullString
		var notes sql.NullString
		var publisher sql.NullString
		var sourceType sql.NullString
		var title sql.NullString
		var url sql.NullString

		err = rows.Scan(&id, &author, &date, &notes, &publisher, &sourceType, &title, &url)
		utils.CheckErr(err)

		source := models.Source{getInt64(id), getString(author), getString(date), getString(notes), getString(publisher), getString(sourceType), getString(title), getString(url)}

		sources = append(sources, source)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(sources)
}

//-----------------------------------------------------------------------------
// sourceCreate
//-----------------------------------------------------------------------------

func sourceCreate(w http.ResponseWriter, r *http.Request) {

	var source models.Source

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &source); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("INSERT INTO source (author, date, notes, publisher, title, type, url) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		source.Author, source.Date, source.Notes, source.Publisher, source.Title, source.SourceType, source.Url)

	utils.CheckErr(err)

	w.WriteHeader(http.StatusCreated)
}

//-----------------------------------------------------------------------------
// sourceUpdate
//-----------------------------------------------------------------------------

func sourceUpdate(w http.ResponseWriter, r *http.Request) {

	var source models.Source

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	utils.CheckErr(err)

	if err := json.Unmarshal(body, &source); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	db := openDB()
	defer db.Close()

	_, err = db.Exec("UPDATE source SET author=$1, date=$2, notes=$3, publisher=$4, title=$5, type=$6, url=$7 WHERE id=$8",
		source.Author, source.Date, source.Notes, source.Publisher, source.Title, source.SourceType, source.Url, source.Id)

	utils.CheckErr(err)
}

//-----------------------------------------------------------------------------
// getCensusHouseholds
//-----------------------------------------------------------------------------

func getCensusHouseholds(censusHouseholdId int64) []models.CensusHousehold {

	db := openDB()
	defer db.Close()

	var rows *sql.Rows
	var err error

	if censusHouseholdId == 0 {
		rows, err = db.Query("SELECT id, address, censusid, citationid, folio, notes, page, piece FROM censushousehold ORDER BY id")
	} else {
		rows, err = db.Query("SELECT id, address, censusid, citationid, folio, notes, page, piece FROM censushousehold WHERE id = $1 ORDER BY id", censusHouseholdId)
	}

	utils.CheckErr(err)
	defer rows.Close()

	censusHouseholds := []models.CensusHousehold{}

	for rows.Next() {
		var id sql.NullInt64
		var address sql.NullString
		var censusId sql.NullInt64
		var citationId sql.NullInt64
		var folio sql.NullString
		var notes sql.NullString
		var page sql.NullString
		var piece sql.NullString

		err = rows.Scan(&id, &address, &censusId, &citationId, &folio, &notes, &page, &piece)
		utils.CheckErr(err)

		censusHousehold := models.CensusHousehold{getInt64(id), getString(address), getInt64(censusId), getInt64(citationId), getString(folio), getString(notes), getString(page), getString(piece), []models.CensusHouseholdPerson{}}

		censusHouseholds = append(censusHouseholds, censusHousehold)
	}

	for index, censusHousehold := range censusHouseholds {
		rows, err = db.Query("SELECT id, age, birthplace, censushouseholdid, name, occupation, personid, relationshiptohead, rownumber, status FROM censushouseholdperson WHERE censushouseholdid = $1 ORDER BY rownumber, id", censusHousehold.Id)

		utils.CheckErr(err)

		persons := []models.CensusHouseholdPerson{}

		for rows.Next() {
			var id sql.NullInt64
			var age sql.NullString
			var birthplace sql.NullString
			var censusHouseholdId sql.NullInt64
			var name sql.NullString
			var occupation sql.NullString
			var personid sql.NullInt64
			var relationshipToHead sql.NullString
			var rownumber sql.NullInt64
			var status sql.NullString

			err = rows.Scan(&id, &age, &birthplace, &censusHouseholdId, &name, &occupation, &personid, &relationshipToHead, &rownumber, &status)
			utils.CheckErr(err)

			censusHouseholdPerson := models.CensusHouseholdPerson{getInt64(id), getString(age), getString(birthplace), getInt64(censusHouseholdId), getString(name), getString(occupation), getInt64(personid), getString(relationshipToHead), getInt64(rownumber), getString(status)}

			persons = append(persons, censusHouseholdPerson)
		}

		censusHouseholds[index].Persons = persons
	}

	return censusHouseholds
}

//-----------------------------------------------------------------------------
// getBool
//-----------------------------------------------------------------------------

func getBool(value sql.NullBool) bool {
	if value.Valid == true {
		return value.Bool
	} else {
		return false
	}
}

//-----------------------------------------------------------------------------
// getInt64
//-----------------------------------------------------------------------------

func getInt64(value sql.NullInt64) int64 {
	if value.Valid == true {
		return value.Int64
	} else {
		return 0
	}
}

//-----------------------------------------------------------------------------
// getPersonFromRow
//-----------------------------------------------------------------------------

func getPersonFromRow(rows *sql.Rows) models.Person {

	var id sql.NullInt64
	var fatherId sql.NullInt64
	var forenames sql.NullString
	var gender sql.NullString
	var motherId sql.NullInt64
	var notes sql.NullString
	var surname sql.NullString

	err := rows.Scan(&id, &fatherId, &forenames, &gender, &motherId, &notes, &surname)
	utils.CheckErr(err)

	return models.Person{getInt64(id), getInt64(fatherId), getString(forenames), getString(gender), getInt64(motherId), getString(notes), getString(surname)}
}

//-----------------------------------------------------------------------------
// getString
//-----------------------------------------------------------------------------

func getString(value sql.NullString) string {
	if value.Valid == true {
		return value.String
	} else {
		return ""
	}
}

//-----------------------------------------------------------------------------
// openDB
//-----------------------------------------------------------------------------

func openDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s port=5433 sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	utils.CheckErr(err)
	return db
}

//-----------------------------------------------------------------------------
// viewCensusHousehold
//-----------------------------------------------------------------------------

func viewCensusHousehold(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	censusHouseholdId, err := strconv.ParseInt(vars["censusHouseholdId"], 10, 64)
	utils.CheckErr(err)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte("<html>"))
	w.Write([]byte("<head>"))

	writeStylesheet(w)

	w.Write([]byte("</head>"))
	w.Write([]byte("<body>"))

	w.Write([]byte("<h1>Census Household</h1>"))

	censusHousehold := getCensusHousehold(censusHouseholdId)
	censuses := getCensuses()

	var census *models.Census = nil
	for _, _census := range censuses {
		if censusHousehold.CensusId == _census.Id {
			census = &_census
			break
		}
	}

	w.Write([]byte("<table width='50%'>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<td class='header' width='2	0%'>Census</td>"))
	w.Write([]byte("<td>" + census.Title + "</td>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<td class='header'>Address</td>"))
	w.Write([]byte("<td>" + censusHousehold.Address + "</th>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<td class='header'>Reference</td>"))
	w.Write([]byte("<td>" + censusHousehold.Piece + "/" + censusHousehold.Folio + "/" + censusHousehold.Page + "</th>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<td class='header'>Notes</td>"))
	w.Write([]byte("<td>" + censusHousehold.Notes + "</th>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("</table>"))

	w.Write([]byte("<br />"))

	w.Write([]byte("<table width='100%'>"))
	w.Write([]byte("<thead>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<th>Id</th>"))
	w.Write([]byte("<th>Name</th>"))
	w.Write([]byte("<th>Rel</th>"))
	w.Write([]byte("<th>Age</th>"))
	w.Write([]byte("<th>Status</th>"))
	w.Write([]byte("<th>Occupation</th>"))
	w.Write([]byte("<th>Birthplace</th>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("</thead>"))
	w.Write([]byte("<tr>"))

	for _, person := range censusHousehold.Persons {
		w.Write([]byte("<tr>"))
		w.Write([]byte("<td>" + fmt.Sprintf("<a href='http://localhost:8880/view/censushouseholdperson/%d'>%d</a>", person.Id, person.Id) + "</td>"))
		if person.PersonId > 0 {    
			w.Write([]byte(fmt.Sprintf("<td><a href='http://localhost:8880/view/person/%d'>%s</td>", person.PersonId, person.Name)))
		} else {
			w.Write([]byte("<td>" + person.Name + "</td>"))
		}
		w.Write([]byte("<td>" + person.RelationshipToHead + "</td>"))
		w.Write([]byte("<td>" + person.Age + "</td>"))
		w.Write([]byte("<td>" + person.Status + "</td>"))
		w.Write([]byte("<td>" + person.Occupation + "</td>"))
		w.Write([]byte("<td>" + person.Birthplace + "</td>"))
		w.Write([]byte("</tr>"))
	}

	w.Write([]byte("</table>"))

	w.Write([]byte("</body>"))
	w.Write([]byte("</html>"))
}

//-----------------------------------------------------------------------------
// viewPersonAncestors
//-----------------------------------------------------------------------------

func viewPersonAncestors(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	thisPerson := getPersonById(personId)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte("<html>"))
	w.Write([]byte("<head>"))

	writeStylesheet(w)

	w.Write([]byte("</head>"))
	w.Write([]byte("<body>"))

	if thisPerson != nil {
		w.Write([]byte(fmt.Sprintf("<h1>%s</h1>", thisPerson.GetFullName())))

		generation := 0

		nextGenerationIds := []int64{}

		if thisPerson.FatherId > 0 {
			nextGenerationIds = append(nextGenerationIds, thisPerson.FatherId)
		}

		if thisPerson.MotherId > 0 {
			nextGenerationIds = append(nextGenerationIds, thisPerson.MotherId)
		}

		for len(nextGenerationIds) > 0 {
			generation++

			w.Write([]byte(fmt.Sprintf("<h2>%d</h2>", generation)))

			thisGenerationIds := []int64{}

			for _, id := range nextGenerationIds {
				thisGenerationIds = append(thisGenerationIds, id)
			}

			nextGenerationIds = nextGenerationIds[:0]

			for _, id := range thisGenerationIds {
				person := getPersonById(id)

				w.Write([]byte(fmt.Sprintf("%s<br />", person.GetFullName())))

				if person.FatherId > 0 {
					nextGenerationIds = append(nextGenerationIds, person.FatherId)
				}
				if person.MotherId > 0 {
					nextGenerationIds = append(nextGenerationIds, person.MotherId)
				}
			}

		}
	} else {
		w.Write([]byte("<h2>Person not found</h2>"))
	}

	w.Write([]byte("</body>"))
	w.Write([]byte("</html>"))
}

//-----------------------------------------------------------------------------
// viewPerson
//-----------------------------------------------------------------------------

func viewPerson(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	personId, err := strconv.ParseInt(vars["personId"], 10, 64)
	utils.CheckErr(err)

	thisPerson := getPersonById(personId)

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte("<html>"))
	w.Write([]byte("<head>"))

	writeStylesheet(w)

	w.Write([]byte("</head>"))
	w.Write([]byte("<body>"))

	if thisPerson != nil {
		w.Write([]byte(fmt.Sprintf("<h1>%s</h1>", thisPerson.GetFullName())))

		// Parents

		w.Write([]byte("<h2>Parents</h2>"))

		father := getPersonById(thisPerson.FatherId)
		mother := getPersonById(thisPerson.MotherId)

		var fatherBirth *models.Event= nil
		var fatherDeath *models.Event= nil
		var motherBirth *models.Event= nil
		var motherDeath *models.Event= nil

		if father != nil {
			fatherBirth = getPrimaryEvent(father.Id, "Birth")
			if fatherBirth == nil {
				fatherBirth = getPrimaryEvent(father.Id, "Baptism")
			}

			fatherDeath = getPrimaryEvent(father.Id, "Death")
			if fatherDeath == nil {
				fatherDeath = getPrimaryEvent(father.Id, "Burial")
			}
		}

		if mother != nil {
			motherBirth = getPrimaryEvent(mother.Id, "Birth")
			if motherBirth == nil {
				motherBirth = getPrimaryEvent(mother.Id, "Baptism")
			}

			motherDeath = getPrimaryEvent(mother.Id, "Death")
			if motherDeath == nil {
				motherDeath = getPrimaryEvent(mother.Id, "Burial")
			}
		}

		w.Write([]byte("<table width='75%'>"))
		w.Write([]byte("<thead>"))
		w.Write([]byte("<tr>"))
		w.Write([]byte("<th>Id</th>"))
		w.Write([]byte("<th>Name</th>"))
		w.Write([]byte("<th>Birth</th>"))
		w.Write([]byte("<th>Death</th>"))
		w.Write([]byte("</tr>"))
		w.Write([]byte("</thead>"))
		w.Write([]byte("<tr>"))
		if father != nil {
			w.Write([]byte("<td width='10%'>"))
			w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/person/%d'>%d</a>", thisPerson.FatherId, thisPerson.FatherId)))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte(father.GetFullName()))
			w.Write([]byte("</td>"))
			if fatherBirth != nil {
				w.Write([]byte("<td width='30%'>" + fmt.Sprintf("%s %s</td>", fatherBirth.Date, fatherBirth.Location)))
			} else {
				w.Write([]byte("<td width='30%'>&nbsp;</td>"))
			}
			if fatherDeath != nil {
				w.Write([]byte("<td width='30%'>" + fmt.Sprintf("%s %s</td>", fatherDeath.Date, fatherDeath.Location)))
			} else {
				w.Write([]byte("<td width='30%'>&nbsp;</td>"))
			}
		} else {
			w.Write([]byte("<td width='10%'>&nbsp;</td>"))
			w.Write([]byte("<td />"))
			w.Write([]byte("<td>"))
			w.Write([]byte("&nbsp;"))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte("&nbsp;"))
			w.Write([]byte("</td>"))
		}
		w.Write([]byte("</tr>"))
		w.Write([]byte("<tr>"))
		if mother != nil {
			w.Write([]byte("<td width='10%'>"))
			w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/person/%d'>%d</a>", thisPerson.MotherId, thisPerson.MotherId)))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte(mother.GetFullName()))
			w.Write([]byte("</td>"))
			if motherBirth != nil {
				w.Write([]byte(fmt.Sprintf("<td>%s %s</td>", motherBirth.Date, motherBirth.Location)))
			} else {
				w.Write([]byte("<td>&nbsp;</td>"))
			}
			if motherDeath != nil {
				w.Write([]byte(fmt.Sprintf("<td>%s %s</td>", motherDeath.Date, motherDeath.Location)))
			} else {
				w.Write([]byte("<td>&nbsp;</td>"))
			}
		} else {
			w.Write([]byte("<td width='10%'>&nbsp;</td>"))
			w.Write([]byte("<td />"))
			w.Write([]byte("<td>"))
			w.Write([]byte("&nbsp;"))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte("&nbsp;"))
			w.Write([]byte("</td>"))
		}
		w.Write([]byte("</tr>"))
		w.Write([]byte("</table>"))

		// Timeline

		w.Write([]byte("<h2>Timeline</h2>"))

		events := getEventsForPerson(personId)
		children := getChildrenForPerson(personId)
		marriages := getMarriagesForPerson(personId)
		censuses := getCensuses()
		censusHouseholds := getCensusHouseholdsForPerson(personId)

		timelineEvents := []models.TimelineEvent{}

		for _, event := range events {
			timelineEvents = append(timelineEvents, models.TimelineEvent{event.Id, event.Date, event.Details, event.EventType, event.IsPrimary, event.Location})
		}

		for _, child := range children {
			childBirth := getPrimaryEvent(child.Id, "Birth")
			if childBirth != nil {
				timelineEvents = append(timelineEvents, models.TimelineEvent{child.Id, childBirth.Date, child.GetFullName(), "Child Born", childBirth.IsPrimary, childBirth.Location})
			}
		}

		for _, marriage := range marriages {
			var spouse *models.Person = nil
			if marriage.HusbandId == personId {
				spouse = getPersonById(marriage.WifeId)
			} else {
				spouse = getPersonById(marriage.HusbandId)
			}

			if spouse != nil {
				timelineEvents = append(timelineEvents, models.TimelineEvent{marriage.Id, marriage.Date, spouse.GetFullName(), "Marriage", false, marriage.Location})
			}
		}

		for _, census := range censuses {
			var censusHousehold *models.CensusHousehold = nil
			for _, _censusHousehold := range censusHouseholds {
				if _censusHousehold.CensusId == census.Id {
					censusHousehold = &_censusHousehold
					break
				}
			}

			if censusHousehold != nil {
				timelineEvents = append(timelineEvents, models.TimelineEvent{censusHousehold.Id, census.Date, "", "Census", false, censusHousehold.Address})
			}
		}

		sort.Sort(models.ByDate(timelineEvents))

		w.Write([]byte("<table width='100%'>"))
		w.Write([]byte("<thead>"))
		w.Write([]byte("<tr>"))
		w.Write([]byte("<th width='7%'>Id</th>"))
		w.Write([]byte("<th width='10%'>Event</th>"))
		w.Write([]byte("<th width='13%'>Date</th>"))
		w.Write([]byte("<th width='40%'>Localtion</th>"))
		w.Write([]byte("<th width='30%'>Details</th>"))
		w.Write([]byte("</tr>"))
		w.Write([]byte("</thead>"))

		for _, timelineEvent := range timelineEvents {
			w.Write([]byte("<tr>"))
			w.Write([]byte("<td>"))
			
			if timelineEvent.EventType == "Census" {
				w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/censushousehold/%d'>%d</a>", timelineEvent.Id, timelineEvent.Id)))
			} else if timelineEvent.EventType == "Marriage" {
				w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/marriage/%d'>%d</a>", timelineEvent.Id, timelineEvent.Id)))
			} else if timelineEvent.EventType == "Child Born" {
				w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/person/%d'>%d</a>", timelineEvent.Id, timelineEvent.Id)))
			} else {
				w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/event/%d'>%d</a>", timelineEvent.Id, timelineEvent.Id)))
			}
			
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte(timelineEvent.EventType))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte(timelineEvent.Date))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte(timelineEvent.Location))
			w.Write([]byte("</td>"))
			w.Write([]byte("<td>"))
			w.Write([]byte(timelineEvent.Details))
			w.Write([]byte("</td>"))
			w.Write([]byte("</tr>"))
		}

		w.Write([]byte("</table>"))

		// Census

		w.Write([]byte("<h2>Census</h2>"))

		w.Write([]byte("<table width='100%'>"))
		w.Write([]byte("<thead>"))
		w.Write([]byte("<tr>"))
		w.Write([]byte("<th width='7%'>Id</th>"))
		w.Write([]byte("<th>Date</th>"))
		w.Write([]byte("<th>Census</th>"))
		w.Write([]byte("<th>Address</th>"))
		w.Write([]byte("<th>Name</th>"))
		w.Write([]byte("<th>Age</th>"))
		w.Write([]byte("<th>Occupation</th>"))
		w.Write([]byte("</tr>"))
		w.Write([]byte("</thead>"))

		for _, census := range censuses {
			var censusHousehold *models.CensusHousehold = nil
			for _, _censusHousehold := range censusHouseholds {
				if _censusHousehold.CensusId == census.Id {
					censusHousehold = &_censusHousehold
					break
				}
			}

			if censusHousehold != nil {
				var person *models.CensusHouseholdPerson = nil

				for _, _person := range censusHousehold.Persons {
					if _person.PersonId == personId {
						person = &_person
						break
					}
				}

				if person != nil {
					w.Write([]byte("<tr>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(fmt.Sprintf("<a href='http://localhost:8880/view/censushousehold/%d'>%d</a>", censusHousehold.Id, censusHousehold.Id)))
					w.Write([]byte("</td>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(census.Date))
					w.Write([]byte("</td>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(census.Title))
					w.Write([]byte("</td>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(censusHousehold.Address))
					w.Write([]byte("</td>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(person.Name))
					w.Write([]byte("</td>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(person.Age))
					w.Write([]byte("</td>"))
					w.Write([]byte("<td>"))
					w.Write([]byte(person.Occupation))
					w.Write([]byte("</td>"))
					w.Write([]byte("</tr>"))
				}
			}
		}

		w.Write([]byte("</table>"))

		// Marriages

		w.Write([]byte("<h2>Marriages</h2>"))

		w.Write([]byte("<table width='75%'>"))
		w.Write([]byte("<thead>"))
		w.Write([]byte("<tr>"))
		w.Write([]byte("<th width='10%'>Id</th>"))
		w.Write([]byte("<th width='20%'>Date</th>"))
		w.Write([]byte("<th width='30%'>Spouse</th>"))
		w.Write([]byte("<th width='40%'>Localtion</th>"))
		w.Write([]byte("</tr>"))
		w.Write([]byte("</thead>"))

		for _, marriage := range marriages {
			var spouse *models.Person = nil
			if marriage.HusbandId == personId {
				spouse = getPersonById(marriage.WifeId)
			} else {
				spouse = getPersonById(marriage.HusbandId)
			}

			w.Write([]byte("<tr>"))
			w.Write([]byte(fmt.Sprintf("<td><a href='http://localhost:8880/view/marriage/%d'>%d</a></td>", marriage.Id, marriage.Id)))
			w.Write([]byte(fmt.Sprintf("<td>%s</td>", marriage.Date)))
			w.Write([]byte(fmt.Sprintf("<td><a href='http://localhost:8880/view/person/%d'>%s</a></td>", spouse.Id, spouse.GetFullName())))
			w.Write([]byte(fmt.Sprintf("<td>%s</td>", marriage.Location)))
			w.Write([]byte("</tr>"))
		}

		w.Write([]byte("</table>"))

		// Children

		sort.Sort(models.ByName(children))

		w.Write([]byte("<h2>Children</h2>"))

		w.Write([]byte("<table width='75%'>"))
		w.Write([]byte("<thead>"))
		w.Write([]byte("<tr>"))
		w.Write([]byte("<th width='10%'>Id</th>"))
		w.Write([]byte("<th>Name</th>"))
		w.Write([]byte("<th width='30%'>Birth</th>"))
		w.Write([]byte("<th width='30%'>Death</th>"))
		w.Write([]byte("</tr>"))
		w.Write([]byte("</thead>"))

		for _, child := range children {
			var birth *models.Event= getPrimaryEvent(child.Id, "Birth")
			if birth == nil {
				birth = getPrimaryEvent(child.Id, "Baptism")
			}

			var death *models.Event= getPrimaryEvent(child.Id, "Death")
			if death == nil {
				death = getPrimaryEvent(child.Id, "Burial")
			}

			w.Write([]byte("<tr>"))
			w.Write([]byte(fmt.Sprintf("<td><a href='http://localhost:8880/view/person/%d'>%d</a></td>", child.Id, child.Id)))
			w.Write([]byte(fmt.Sprintf("<td>%s</td>", child.GetFullName())))
			if birth != nil {
				w.Write([]byte(fmt.Sprintf("<td>%s %s</td>", birth.Date, birth.Location)))
			} else {
				w.Write([]byte("<td>&nbsp;</td>"))
			}
			if death != nil {
				w.Write([]byte(fmt.Sprintf("<td>%s %s</td>", death.Date, death.Location)))
			} else {
				w.Write([]byte("<td>&nbsp;</td>"))
			}
			w.Write([]byte("</tr>"))
		}

		w.Write([]byte("</table>"))

		w.Write([]byte("<br />"))
	} else {
		w.Write([]byte("<h2>Person not found</h2>"))
	}

	w.Write([]byte("</body>"))
	w.Write([]byte("</html>"))
}

//-----------------------------------------------------------------------------
// viewPersons
//-----------------------------------------------------------------------------

func viewPersons(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	surname := vars["surname"]

	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte("<html>"))
	w.Write([]byte("<head>"))

	writeStylesheet(w)

	w.Write([]byte("</head>"))
	w.Write([]byte("<body>"))

	w.Write([]byte(fmt.Sprintf("<h1>%s</h1>", surname)))

	w.Write([]byte("<table width='100%'>"))
	w.Write([]byte("<thead>"))
	w.Write([]byte("<tr>"))
	w.Write([]byte("<th width='6%'>Id</th>"))
	w.Write([]byte("<th width='24%'>Name</th>"))
	w.Write([]byte("<th colspan='2' width='35%'>Birth</th>"))
	w.Write([]byte("<th colspan='2' width='35%'>Death</th>"))
	w.Write([]byte("</tr>"))
	w.Write([]byte("</thead>"))

	persons := getPersonsBySurname(surname)
	sort.Sort(models.ByName(persons))

	for _, person := range persons {
		var birth *models.Event= getPrimaryEvent(person.Id, "Birth")
		if birth == nil {
			birth = getPrimaryEvent(person.Id, "Baptism")
		}

		var death *models.Event= getPrimaryEvent(person.Id, "Death")
		if death == nil {
			death = getPrimaryEvent(person.Id, "Burial")
		}

		w.Write([]byte("<tr>"))
		w.Write([]byte(fmt.Sprintf("<td><a href='http://localhost:8880/view/person/%d'>%d</a></td>", person.Id, person.Id)))
		w.Write([]byte(fmt.Sprintf("<td>%s</td>", person.GetFullName())))
		if birth != nil {
			w.Write([]byte("<td width='10%'>" + fmt.Sprintf("%s</td>", birth.Date)))
			w.Write([]byte(fmt.Sprintf("<td>%s</td>", birth.Location)))
		} else {
			w.Write([]byte("<td>&nbsp;</td>"))
			w.Write([]byte("<td>&nbsp;</td>"))
		}
		if death != nil {
			w.Write([]byte("<td width='10%'>" + fmt.Sprintf("%s</td>", death.Date)))
			w.Write([]byte(fmt.Sprintf("<td>%s</td>", death.Location)))
		} else {
			w.Write([]byte("<td>&nbsp;</td>"))
			w.Write([]byte("<td>&nbsp;</td>"))
		}
		w.Write([]byte("</tr>"))
	}

	w.Write([]byte("</table>"))

	w.Write([]byte("<br />"))

	w.Write([]byte("</body>"))
	w.Write([]byte("</html>"))
}

//-----------------------------------------------------------------------------
// getEventsForPerson
//-----------------------------------------------------------------------------

func getEventsForPerson(personId int64) []models.Event{

	db := openDB()

	rows, err := db.Query("SELECT id, citationid, date, details, isprimary, location, notes, personid, type FROM event WHERE personid = $1 ORDER BY id", personId)
	utils.CheckErr(err)

	events := []models.Event{}

	for rows.Next() {
		var id sql.NullInt64
		var citationId sql.NullInt64
		var date sql.NullString
		var details sql.NullString
		var eventType sql.NullString
		var isPrimary sql.NullBool
		var location sql.NullString
		var notes sql.NullString
		var personId sql.NullInt64

		err = rows.Scan(&id, &citationId, &date, &details, &isPrimary, &location, &notes, &personId, &eventType)
		utils.CheckErr(err)

		event := models.Event{getInt64(id), getInt64(citationId), getString(date), getString(details), getString(eventType), getBool(isPrimary), getString(location), getString(notes), getInt64(personId)}

		events = append(events, event)
	}

	rows.Close()
	db.Close()

	return events
}

//-----------------------------------------------------------------------------
// getChildrenForPerson
//-----------------------------------------------------------------------------

func getChildrenForPerson(personId int64) []models.Person {

	db := openDB()

	rows, err := db.Query("SELECT id, fatherid, forenames, gender, motherid, notes, surname FROM person WHERE fatherid = $1 OR motherId = $1 ORDER BY id", personId)
	utils.CheckErr(err)

	persons := []models.Person{}

	for rows.Next() {
		persons = append(persons, getPersonFromRow(rows))
	}

	rows.Close()
	db.Close()

	return persons
}

//-----------------------------------------------------------------------------
// getEventsForPerson
//-----------------------------------------------------------------------------

func getMarriagesForPerson(personId int64) []models.Marriage {

	db := openDB()

	rows, err := db.Query("SELECT id, citationid, date, husbandid, location, notes, wifeid FROM marriage WHERE husbandid = $1 OR wifeid = $1 ORDER BY id", personId)
	utils.CheckErr(err)

	marriages := []models.Marriage{}

	for rows.Next() {
		var id sql.NullInt64
		var citationId sql.NullInt64
		var date sql.NullString
		var husbandId sql.NullInt64
		var location sql.NullString
		var notes sql.NullString
		var wifeId sql.NullInt64

		err = rows.Scan(&id, &citationId, &date, &husbandId, &location, &notes, &wifeId)
		utils.CheckErr(err)

		marriage := models.Marriage{getInt64(id), getInt64(citationId), getString(date), getInt64(husbandId), getString(location), getString(notes), getInt64(wifeId)}

		marriages = append(marriages, marriage)
	}

	rows.Close()
	db.Close()

	return marriages
}

//-----------------------------------------------------------------------------
// getPersonById
//-----------------------------------------------------------------------------

func getPersonById(personId int64) *models.Person {

	db := openDB()

	rows, err := db.Query("SELECT id, fatherid, forenames, gender, motherid, notes, surname FROM person WHERE id = $1", personId)
	utils.CheckErr(err)

	var person *models.Person = nil
	if rows.Next() {
		_person := getPersonFromRow(rows)
		person = &_person
	}

	rows.Close()
	db.Close()

	return person
}

//-----------------------------------------------------------------------------
// getPersonsBySurname
//-----------------------------------------------------------------------------

func getPersonsBySurname(surname string) []models.Person {

	db := openDB()

	rows, err := db.Query("SELECT id, fatherid, forenames, gender, motherid, notes, surname FROM person WHERE surname = $1", surname)
	utils.CheckErr(err)

	persons := []models.Person{}

	for rows.Next() {
		person := getPersonFromRow(rows)
		persons = append(persons, person)
	}

	rows.Close()
	db.Close()

	return persons
}

//-----------------------------------------------------------------------------
// getCensusHousehold
//-----------------------------------------------------------------------------

func getCensusHousehold(censusHouseholdId int64) models.CensusHousehold {

	return getCensusHouseholds(censusHouseholdId)[0]
}

//-----------------------------------------------------------------------------
// getCensusHouseholdsForPerson
//-----------------------------------------------------------------------------

func getCensusHouseholdsForPerson(personId int64) []models.CensusHousehold {

	db := openDB()

	rows, err := db.Query("SELECT censushouseholdid FROM censushouseholdperson WHERE personid = $1 ORDER BY id", personId)
	utils.CheckErr(err)

	censusHouseholdIds := []int64{}

	for rows.Next() {
		var censusHouseholdId sql.NullInt64

		err = rows.Scan(&censusHouseholdId)
		utils.CheckErr(err)

		censusHouseholdIds = append(censusHouseholdIds, getInt64(censusHouseholdId))
	}

	censusHouseholds := []models.CensusHousehold{}

	for _, id := range censusHouseholdIds {
		censusHouseholds = append(censusHouseholds, getCensusHouseholds(id)[0])
	}

	rows.Close()
	db.Close()

	return censusHouseholds
}

//-----------------------------------------------------------------------------
// getCensuses
//-----------------------------------------------------------------------------

func getCensuses() []models.Census {

	db := openDB()

	rows, err := db.Query("SELECT id, date, title FROM census ORDER BY id")
	utils.CheckErr(err)

	censuses := []models.Census{}

	for rows.Next() {
		var id sql.NullInt64
		var date sql.NullString
		var title sql.NullString

		err = rows.Scan(&id, &date, &title)
		utils.CheckErr(err)

		census := models.Census{getInt64(id), getString(date), getString(title)}

		censuses = append(censuses, census)
	}

	rows.Close()
	db.Close()

	return censuses
}

//-----------------------------------------------------------------------------
// writeStylesheet
//-----------------------------------------------------------------------------

func writeStylesheet(w http.ResponseWriter) {

	w.Write([]byte("<style type='text/css'>"))
	w.Write([]byte("body {"))
	w.Write([]byte("color: #000000;"))
	w.Write([]byte("background-color: #ffffff;"))
	w.Write([]byte("font-family: consolas, arial;"))
	w.Write([]byte("}"))
	w.Write([]byte("table {"))
	w.Write([]byte("border-style: solid;"))
	w.Write([]byte("border-color: #000000;"))
	w.Write([]byte("border-collapse: collapse;"))
	w.Write([]byte("cursor: pointer;"))
	w.Write([]byte("}"))
	w.Write([]byte("th, .header {"))
	w.Write([]byte("background: #cccccc;"))
	w.Write([]byte("font-weight: bold;"))
	w.Write([]byte("cursor: pointer;"))
	w.Write([]byte("}"))
	w.Write([]byte("td, th {"))
	w.Write([]byte("border-style: solid;"))
	w.Write([]byte("border-width: thin;"))
	w.Write([]byte("padding-left: 1em;"))
	w.Write([]byte("padding-right: 1em;"))
	w.Write([]byte("padding-top: 0.5em;"))
	w.Write([]byte("padding-bottom: 0.5em;"))
	w.Write([]byte("}"))
	w.Write([]byte("</style>"))
}

//-----------------------------------------------------------------------------
// getPrimaryEvent
//-----------------------------------------------------------------------------

func getPrimaryEvent(personId int64, eventType string) *models.Event{

	db := openDB()

	rows, err := db.Query("SELECT id, citationid, date, details, isprimary, location, notes, personid, type FROM event WHERE personid = $1 AND type = $2 AND isprimary ORDER BY id", personId, eventType)
	utils.CheckErr(err)

	var event *models.Event= nil

	if rows.Next() {
		event = getEventFromRow(rows)
	} else {
		rows, err = db.Query("SELECT id, citationid, date, details, isprimary, location, notes, personid, type FROM event WHERE personid = $1 AND type = $2 ORDER BY id", personId, eventType)
		utils.CheckErr(err)
		if rows.Next() {
			event = getEventFromRow(rows)
		}
	}

	rows.Close()
	db.Close()

	return event
}

//-----------------------------------------------------------------------------
// getEventFromRow
//-----------------------------------------------------------------------------

func getEventFromRow(rows *sql.Rows) *models.Event{

	var id sql.NullInt64
	var citationId sql.NullInt64
	var date sql.NullString
	var details sql.NullString
	var eventType sql.NullString
	var isPrimary sql.NullBool
	var location sql.NullString
	var notes sql.NullString
	var personId sql.NullInt64

	err := rows.Scan(&id, &citationId, &date, &details, &isPrimary, &location, &notes, &personId, &eventType)
	utils.CheckErr(err)

	return &models.Event{getInt64(id), getInt64(citationId), getString(date), getString(details), getString(eventType), getBool(isPrimary), getString(location), getString(notes), getInt64(personId)}
}
