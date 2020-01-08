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

In the first part of the function, process the POST request. This is the data sent by the client in the page request:
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


## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
