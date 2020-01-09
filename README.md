# GoWebTable

Go library to dynamically create interactive database-based data tables for Go (Golang) web applications.

![Data Table Screenshot](/img/Screenshot1.png?raw=true "Data Table Screenshot")

GoWebTables feature ordering, pagination and (currently) filtering on a pre-specified field.

## Getting Started

### Prerequisites

Any recent installation of Go should be sufficient.

For the front end table interactivity, JQuery is used for various functions and will therefore need to be available locally or by CDN.

In the example below, Gorilla Mux is used for the web server. This is not essential but highly recommended.

### Installing

To install the go library:
```
go get -u github.com/fishboy25uk/GoWebTable
```

## Deployment
First, include the library in your project
```
import gwt "github.com/fishboy25uk/gowebtable"
```
Create a Go web server. Here Gorilla Mux is being used, but this is optional:
```
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handlerIndex)
	r.HandleFunc("/data/", handlerData)

	log.Fatal(http.ListenAndServe(":80", r))
}
```
Create a handler function for the initial web page:
```
func handlerIndex(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	info := make(map[string]interface{})

	t := template.Must(template.ParseFiles("index.html"))
	err := t.ExecuteTemplate(w, "index.html", &info)
	if err != nil {
		log.Printf("ERROR: handlerIndex ExecuteTemplate - %s\n", err)
	}

}
```
Create a HTML template which has a target DIV for where the table will go. This will be loaded into the DIV on page load.
Note that jQuery added in the head which is currently required for frontend function
```
<!doctype html>
<html lang="en">

<head>
    <script src="https://code.jquery.com/jquery-3.4.1.min.js"></script>
</head>

<body>
    <div id="target"></div>
</body>
<script>
    $('#target').load('/data/');
</script>
</html>
```
Now create a handler function for the data table loader:
```
func handlerData(w http.ResponseWriter, r *http.Request) {

}
```
In the first part of this handler function, create a Fields slice to define the fields used for the data table
```
//Define fields for table
fields := []gwt.TableField{
	{Name: "id", Title: "ID", DBName: "id", Type: "string"},
	{Name: "name", Title: "Name", DBName: "name", Type: "string"},
	{Name: "type", Title: "Type", DBName: "type", Type: "string"},
}
```
Now create a PageDetails object then process the POST request - this is the data sent by the client in the page request.
The POST data will be unmarshalled into the pd (PageDetails) object.
```
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
```
Now set the default PageDetails options; URL location, the target DIV name, the default order element name and the default filter box element i.e. the element the filter term will act on.
```
//Set PageDetails object parameters
pd.URL = "/data/"
pd.Target = "target"
pd.OrderDefaultElement = "name"
pd.FilterDefaultElement = "name"
```
Launch the PreCalculate function, which sets options used in the subsequent database queries:
```
//PreCalculate
pd.PreCalculate()
```
Now get the count of the total records from the database. This will be your own function which performs a SELECT COUNT(\*) from the database. In this example the function is called selectRecordsTotalAll(). Add the total to TotalAll of the Page Details object:
```
//Get all records total
totalAll, err := selectRecordsTotalAll()
//Handle error
pd.TotalAll = totalAll
```
Next we get the total for records but with filter applied, if filters are active, and set the returned total to the TotalFiltered field of the Page Details object. If no filter terms are specified in the request, the TotalFiltered field is set to the same as Total
```
//Get filtered records total
if len(pd.FilterTerms) > 0 {
	totalFiltered, err := selectRecordsTotalFiltered(&pd)
	//Handle error

	pd.TotalFiltered = totalFiltered
	pd.IsFiltered = true
} else {
	pd.TotalFiltered = totalAll
}
```
Now we know the total records we can calculate what offset is required for the main SQL query by running the Calculate function:
```
pd.Calculate()
```
Now get the main records from the database. Here we pass the Page Details object to the function so it can build a custom SQL query with the offset and limit we need, as well as any order and filter options specified. See example_test.go for an example of a function (selectRecords) to build the SQL string, do the SQL query and return the records.
```
//Get records
records, err := selectRecords(&pd)
//Handle error
}
```
The final table will by dynamically generated from the Fields object and the Records data. The Golang template/html template cannot interate the fields in a struct object (they would normally be explicitly named e.g. {{.Name}}). Therefore, the records data needs to be in the format of a slice of slices of the values e.g. [["1","Foo1","string"],["1","Foo1","string"]]. In this example the faith/structs package is used to perform this conversion.
```
//Convert records into a map slice
var recordsSlice [][]interface{}
for r := range records {
	recordsSlice = append(recordsSlice, structs.Values(records[r]))
}
```
Finally we pass the Page Details object, Fields object and the records (via the 'info' map object) into the GoWebTable HTML template, which is loaded by the getTableTemplate function. Once this is done, execute the template and pass to the writer:
```
info := make(map[string]interface{})
info["PageDetails"] = pd
info["Fields"] = fields
info["Records"] = recordsSlice

t, err := template.New("table").Parse(tableTemplateData)
//Handle error

err = t.Execute(w, &info)
//Handle error
```
The complete data handler function is show below:
```
func handlerData(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	//Define fields for table
	fields := []gwt.TableField{
		{Name: "id", Title: "ID", DBName: "id", Type: "string"},
		{Name: "name", Title: "Name", DBName: "name", Type: "string"},
		{Name: "type", Title: "Type", DBName: "type", Type: "string"},
	}

	//Create PageDetails object
	var pd gwt.PageDetails

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
	pd.FilterDefaultElement = "name"

	//PreCalculate limit
	pd.PreCalculate()

	//Get all records total
	totalAll, err := selectRecordsTotalAll()
	if err != nil {
		log.Printf("ERROR: handlerExampleselectRecordsTotalAll - %s\n", err)
	}
	pd.TotalAll = totalAll

	//Get filtered records total
	if len(pd.FilterTerms) > 0 {

		totalFiltered, err := selectRecordsTotalFiltered(&pd)

		if err != nil {
			log.Printf("ERROR: handlerExampleselectRecordsTotalFiltered - %s\n", err)
		}

		pd.TotalFiltered = totalFiltered
		pd.IsFiltered = true
	} else {
		pd.TotalFiltered = totalAll
	}

	//Calculate parameters
	pd.Calculate()

	//Get records
	records, err := selectRecords(&pd)
	if err != nil {
		log.Printf("ERROR: handlerExample selectRecords - %s\n", err)
	}

	//Convert records into a map slice
	var recordsSlice [][]interface{}
	for r := range records {
		recordsSlice = append(recordsSlice, structs.Values(records[r]))
	}

	info := make(map[string]interface{})
	info["PageDetails"] = pd
	info["Fields"] = fields
	info["Records"] = recordsSlice

	t, err := template.New("table").Parse(tableTemplateData)
	if err != nil {
		log.Printf("ERROR: handlerExample Parse Template - %s\n", err)
	}
	err = t.Execute(w, &info)
	if err != nil {
		log.Printf("ERROR: handlerExample Execute Template - %s\n", err)
	}

}
```
## Example
See example_test.go for a full example of how to create a data table. This example uses CrateDB for the database, but any MySQL compatible DB should be usable.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
