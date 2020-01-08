# GoWebTable

Go library to help create interactive DB-based data tables for web applications written in Go (Golang)

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
First, create a Go web server:
```
func main() {

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", handlerIndex)
	r.HandleFunc("/data/", handlerData)

	// Bind to a port and pass our router in
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
Create a HTML template which has a target DIV for where the table will go. This will be loaded into the DIV on page load:
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
Now create a handler function for the table loader:

```
func handlerData(w http.ResponseWriter, r *http.Request) {

}
```

In the first part of the function, create the PageDetails object then process the POST request. This is the data sent by the client in the page request:
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
Now set the URL location and the target DIV name before launching the PreCalculate function, which sets options for the subsequent database queries:
```
//Set PageDetails object parameters
pd.URL = "/data/"
pd.Target = "target"

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
Next we get the total for records but with filter applied, if filters are active, and setthe returned total to the TotalFiltered field of the Page Details object. If no filter terms are specified in the request, the TotalFiltered field is set to the same as Total
```
//Get filtered records total
if len(pd.FilterTerms) > 0 {

	totalFiltered, err := selectRecordsTotalFiltered(&pd)
	//Handle error

	pd.TotalFiltered = totalFiltered
	pd.IsFiltered = true
}
```

Now we know the total records we can calculate what offset is required for the main SQL query by running the Calculate function:
```
pd.Calculate()
```
Run the main records function. Here we pass the Page Details object to the function so it can build a SQL query with the offset and limit we need, as well as any order and filter options specified.
```
//Get records
records, err := selectRecords(&pd)
//Handle error
}
```
Finally we pass the records and the Page Details object (via the info map object) into the GoWebTable HTML template, which is loaded by the getTableTemplate function. Once this is done, execute the template and pass to the writer:
```
info := make(map[string]interface{})
info["PageDetails"] = pd
info["Records"] = records

t, err := template.New("table").Parse(goTableTemplate)
//Handle error

err = t.Execute(w, &info)
//Handle error
```	

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
