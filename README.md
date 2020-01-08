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
	pd.Table = "tbl"
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
	} else {
		pd.TotalFiltered = pd.TotalAll
		pd.IsFiltered = false
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

	t, err := template.New("table").Parse(goTableTemplate)
	if err != nil {
		log.Printf("ERROR: handlerExample Parse Template - %s\n", err)
	}
	err = t.Execute(w, &info)
	if err != nil {
		log.Printf("ERROR: handlerExample Execute Template - %s\n", err)
	}

}
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
