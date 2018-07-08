package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func main() {
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(bytes.NewReader(content))

	header, _ := r.Read()
	fmt.Printf("# DML\n# CONTEXT-DATABASE: collectd\n\n")

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		r := map[string]string{}
		for i, f := range record {
			r[header[i]] = f
		}

		// Header
		name := r["name"]
		// Tags
		host := r["host"]
		typ := r["type"]
		instance := r["type_instance"]
		// Fields
		value := r["value"]
		v, _ := strconv.ParseFloat(value, 64)
		v -= 9
		// Footer
		time := r["time"]

		fmt.Printf("%s,", name)
		fmt.Printf("host=%s,", host)
		fmt.Printf("type=%s,", typ)
		fmt.Printf("type_instance=%s ", instance)
		fmt.Printf("value=%f ", v)
		fmt.Printf("%s\n", time)
	}
}
