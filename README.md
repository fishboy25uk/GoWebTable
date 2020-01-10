# GoWebTable

Go library to dynamically create interactive database-based data tables for Go (Golang) web applications.

![Data Table Screenshot](/img/Screenshot1.png?raw=true "Data Table Screenshot")

GoWebTables feature ordering, pagination and global term search.

NOTE: THIS PROJECT IS IN DEVELOPMENT

## Getting Started

### Prerequisites

Any recent installation of Go should be sufficient.

For the front end table interactivity, JQuery is currently used for various functions and will therefore need to be available locally or by CDN - this will hopefully be phased out on the future.

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
Define the struct for your database records. The gowebtable tag needs to be included:
```
//Record is an example struct for a record
type Record struct {
	ID   string `gowebtable:"id,ID,false" json:"id"`
	Name string `gowebtable:"name,Name,false" json:"name"`
	Type string `gowebtable:"type,Type,false" json:"type"`
}
```
The gowebtable tag is constructed in the format "NAME,TITLE,HIDE". "NAME" is the name of the field and should match the name of the field in the database. "TITLE" is the format of the field name to be displayed in the table. "HIDE" is whether to hide the field in the table (false by default).

Next, create a Go web server. Here Gorilla Mux is being used, but this is optional:
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
Now create a PageDetails object with the necessary PageDetails options; Table name, URL location, the target DIV name. The default order element name is optional but recommended.
```
//Create PageDetails object
pd := gwt.PageDetails{Table: "records", URL: "/data/", Target: "target"}
pd.OrderElementDefault = "name"
```
Process the POST request - this is the data sent by the client in the page request. The POST data will be unmarshalled into the pd (PageDetails) object.
```
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
Launch the PreProcess function, which sets options used in the subsequent database queries and gets the field details from a specified struct ('Record' in this example):
```
pd.PreProcess(&Record{})
```
Now get the count of the total records from the database. This will be your own function which performs a SELECT COUNT(\*) from the database. In this example the function is called selectRecordsCount() and the PageDetails (pd) object is passed in so that this function can get the filter options and include that in the SQL query. Add the total to RecordsAll of the Page Details object:
```
//Get all records total
resultsCount, err := selectRecordsCount(&pd)
//Handle error
pd.RecordsTotal = resultsCount
```
Now we know the total records we can calculate what offset is required for the main SQL query by running the PageProcess function:
```
pd.PageProcess()
```
Now get the main records from the database. Here we pass the Page Details object to the function so it can build a custom SQL query with the offset and limit we need, as well as any order and filter options specified. See example_test.go for an example of a function (selectRecords) to build the SQL string, do the SQL query and return the records.
```
var resultsRaw []Record
if pd.RecordsTotal == 0 {
	resultsRaw = append(resultsRaw, Record{})
} else {
	resultsRaw, err = selectRecords(&pd)
	//Handle error
}
```
The final table will by dynamically generated from the Fields object and the Records data. The Golang template/html template cannot interate the fields in a struct object (they would normally be explicitly named e.g. {{.Name}}). Therefore, the records data needs to be in the format of a slice of slices of the values e.g. [["1","Foo1","string"],["1","Foo1","string"]]. Run the ResultsProcess function to perform this conversion.
```
pd.ResultsProcess(resultsRaw)
```
Finally we pass the Page Details object into the GoWebTable HTML template, which is loaded by the TemplateGet function - this returns HTML data for the table template, Once this is done, execute the template and pass to the writer:
```
info := make(map[string]interface{})
info["PageDetails"] = pd

t, err := template.New("table").Parse(gwt.TemplateGet())
//Handle error

err = t.Execute(w, &info)
//Handle error
```
## Options
### Custom Limit Options
To specify custom limit options, add the following after creating the PageDetails object (substituting your own values):
```
pd.LimitOptions = []int{2, 5, 10, 20, 30, 50, 100}
```
## Example
See [example_test.go](example_test.go) for a full example of how to create a data table. This example uses CrateDB for the database, but any MySQL compatible DB should be usable.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
