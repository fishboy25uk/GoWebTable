package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	gwt "github.com/fishboy25uk/gowebtable"
	"github.com/gorilla/mux"
	_ "github.com/herenow/go-crate"
)

var (
	crateURL = "http://xxx.xxx.xxx.xxx:xxx"
	db *sql.DB
)

//Record is an example struct for a record
type Record struct {
	ID   string `gowebtable:"id,ID,false" json:"id"`
	Name string `gowebtable:"name,Name,false" json:"name"`
	Type string `gowebtable:"type,Type,false" json:"type"`
}

func init() {
	dbTemp, err := sql.Open("crate", crateURL)
	if err != nil {
		log.Fatal(err)
	}
	db = dbTemp
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	info := make(map[string]interface{})

	t := template.Must(template.ParseFiles("index.html"))
	err := t.ExecuteTemplate(w, "index.html", &info)
	if err != nil {
		log.Printf("ERROR: handlerIndex ExecuteTemplate - %s\n", err)
	}

}

func handlerData(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	//Crate PageDetails object with default limit options
	pd := gwt.PageDetails{Table: "records", URL: "/data/", Target: "target"}

	//Set default order element
	pd.OrderElementDefault = "name"

	//Process POST request
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		err = json.Unmarshal(body, &pd)
		if err != nil {
			log.Printf("ERROR: handlerMain unmarshal %s\n", err)
		}
	}

	//Get struct fields and process page details filters
	pd.PreProcess(&Record{})

	//Get all records total
	resultsCount, err := selectRecordsCount(&pd)
	if err != nil {
		log.Printf("ERROR: handlerExampleselectRecordsTotalAll - %s\n", err)
	}
	pd.RecordsTotal = resultsCount

	pd.PageProcess()

	var resultsRaw []Record
	if pd.RecordsTotal == 0 {
		resultsRaw = append(resultsRaw, Record{})
	} else {
		resultsRaw, err = selectRecords(&pd)
		if err != nil {
			log.Printf("ERROR: handlerData eventsSelect - %s", err)
		}
	}

	//Process Results
	pd.ResultsProcess(resultsRaw)

	//Send PageDetails to template
	info := make(map[string]interface{})
	info["PageDetails"] = pd

	t, err := template.New("table").Parse(gwt.TemplateGet())
	//t, err := template.ParseFiles("table.html")
	if err != nil {
		log.Printf("ERROR: handlerExample Parse Template - %s\n", err)
	}

	err = t.Execute(w, &info)
	if err != nil {
		log.Printf("ERROR: handlerExample Execute Template - %s\n", err)
	}

}

func selectRecords(pd *gwt.PageDetails) ([]Record, error) {

	//Create records slice
	var records []Record

	//Build SQL string
	sql := fmt.Sprintf("SELECT id,name,type FROM records%s%s LIMIT ? OFFSET ?", pd.FilterSQLString, pd.OrderTermsString)

	//Perform SQL query
	rows, err := db.Query(sql, pd.Limit, pd.Offset)
	if err != nil {
		return records, err
	}

	//Iterate returned rows
	for rows.Next() {
		var r Record
		err = rows.Scan(
			&r.ID,
			&r.Name,
			&r.Type)
		if err != nil {
			return records, err
		}
		records = append(records, r)
	}
	rows.Close()

	return records, nil

}

func selectRecordsCount(pd *gwt.PageDetails) (int, error) {
	var total int
	sql := fmt.Sprintf("SELECT COUNT(*) FROM records%s", pd.FilterSQLString)
	err := db.QueryRow(sql).Scan(&total)
	if err != nil {
		return total, err
	}
	return total, nil
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", handlerIndex)
	r.HandleFunc("/data/", handlerData)

	log.Fatal(http.ListenAndServe(":80", r))

}
