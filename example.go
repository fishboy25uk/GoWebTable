package gowebtable

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	_ "github.com/herenow/go-crate"
)

var (
	crateURL = "http://192.168.130.108:4201"

	db *sql.DB
)

//Record is an example struct for a record
type Record struct {
	ID   string
	Name string
	Type string
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

	//Create PageDetails object
	var pd PageDetails

	//Process post data (if present)
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		err = json.Unmarshal(body, &pd)
		if err != nil {
			log.Printf("ERROR: handlerExample Unmarshal PageDetails - %s\n", err)
		}
	}

	//Set PageDetails object parameters
	pd.URL = "/data/"
	pd.Target = "target"
	pd.OrderDefaultElement = "name"

	//PreCalculate limit
	pd.PreCalculate()

	//Get all records total
	totalAll, err := selectRecordsTotalAll()
	if err != nil {
		log.Printf("ERROR: handlerExampleselectRecordsTotalAll - %s\n", err)
	}
	pd.TotalAll = totalAll

	//Get filtered records total - CURRENTLY INACTIVE
	if len(pd.FilterTerms) > 0 {

		totalFiltered, err := selectRecordsTotalFiltered(&pd)

		if err != nil {
			log.Printf("ERROR: handlerExampleselectRecordsTotalFiltered - %s\n", err)
		}

		pd.TotalFiltered = totalFiltered
		pd.IsFiltered = true
	}

	//Calculate parameters
	pd.Calculate()

	//Get records
	records, err := selectRecords(&pd)
	if err != nil {
		log.Printf("ERROR: handlerExample selectRecords - %s\n", err)
	}

	info := make(map[string]interface{})
	info["PageDetails"] = pd
	info["Records"] = records

	//t := template.Must(template.ParseFiles(loadPage))
	//err = t.ExecuteTemplate(w, loadPage, &info)

	t, err := template.New("table").Parse(gowebtableTemplateGet())
	if err != nil {
		log.Printf("ERROR: handlerExample Parse Template - %s\n", err)
	}
	err = t.Execute(w, &info)
	if err != nil {
		log.Printf("ERROR: handlerExample Execute Template - %s\n", err)
	}

}

func selectRecords(pd *PageDetails) ([]Record, error) {

	//Create records slice
	var records []Record

	//Filter Terms
	filterString := ""
	if len(pd.FilterTerms) > 0 {
		for _, ft := range pd.FilterTerms {
			if len(ft.Term) > 0 {

				if len(filterString) == 0 {
					filterString = " WHERE "
				} else {
					filterString += " AND "
				}

				//filterString += fmt.Sprintf("%s LIKE '%%%s%%'", ft.Element, ft.Term)
				filterString += fmt.Sprintf("LOWER(%s) LIKE '%%%s%%'", ft.Element, strings.ToLower(ft.Term))

			}
		}
	}

	//Order
	var ordersString string
	if len(pd.OrderTerms) > 0 {
		var ordersArray []string
		for _, o := range pd.OrderTerms {

			if o.Element == "" {
				continue
			}

			ordersArray = append(ordersArray, fmt.Sprintf("%s %s", o.Element, strings.ToUpper(o.Direction)))

		}
		ordersString = " ORDER BY " + strings.Join(ordersArray, ",")
	}

	//Pagination String
	paginationString := fmt.Sprintf(" LIMIT %v OFFSET %v", pd.Limit, pd.Offset)

	//Build SQL string
	sql := fmt.Sprintf("SELECT id,name,type FROM records%s%s%s", filterString, ordersString, paginationString)

	fmt.Println(sql)

	//Perform SQL query
	rows, err := db.Query(sql)
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

	//Return records
	return records, nil

}

func selectRecordsTotalAll() (int, error) {

	//Build SQL string
	sql := fmt.Sprintf("SELECT COUNT(*) FROM records")
	fmt.Println(sql)

	//Perform SQL query
	var count int
	err := db.QueryRow(sql).Scan(&count)
	if err != nil {
		return count, err
	}

	//Return records
	return count, nil

}

func selectRecordsTotalFiltered(pd *PageDetails) (int, error) {

	//Filter Terms
	filterString := ""
	if len(pd.FilterTerms) > 0 {
		for _, ft := range pd.FilterTerms {
			if len(ft.Term) > 0 {

				if len(filterString) == 0 {
					filterString = " WHERE "
				} else {
					filterString += " AND "
				}

				filterString += fmt.Sprintf("LOWER(%s) LIKE '%%%s%%'", ft.Element, strings.ToLower(ft.Term))
			}
		}
	}

	//Build SQL string
	sql := fmt.Sprintf("SELECT COUNT(*) FROM records%s", filterString)
	fmt.Println(sql)

	//Perform SQL query
	var count int
	err := db.QueryRow(sql).Scan(&count)
	if err != nil {
		return count, err
	}

	//Return records
	return count, nil

}

func main() {

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", handlerIndex)
	r.HandleFunc("/data/", handlerData)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":80", r))

}
