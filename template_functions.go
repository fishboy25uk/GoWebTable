package gowebtable

import (
	"bytes"
	"compress/gzip"
	b64 "encoding/base64"
	"io"
	"io/ioutil"
	"log"
)

//var tableTemplateData string

//TemplateGet returns the HTML table template
func TemplateGet() string {
	sDec, err := b64.StdEncoding.DecodeString(tableTemplateData)
	if err != nil {
		log.Fatal(err)
	}
	var buf2 bytes.Buffer
	err = gunzipWrite(&buf2, sDec)
	if err != nil {
		log.Fatal(err)
	}
	return buf2.String()
}

// Write gunzipped data to a Writer
func gunzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	defer gr.Close()
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}
