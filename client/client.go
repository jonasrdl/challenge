package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	const PORT int = 8080

	method := flag.String("m", "", "Method: put, get, or delete")
	key := flag.String("key", "", "Key")
	value := flag.String("value", "", "Value (only for put)")
	flag.Parse()

	if len(*method) == 0 || len(*key) == 0 {
		printUsage()
		_, err := fmt.Fprintln(os.Stderr, "Error: method and key flags must be provided")
		if err != nil {
			log.Println(err.Error())
			return
		}
		os.Exit(1)
	}

	if *method == "put" && len(*value) == 0 {
		_, err := fmt.Fprintln(os.Stderr, "Error: value flag must be provided for put method")
		if err != nil {
			log.Println(err.Error())
			return
		}
		os.Exit(1)
	}

	var url string
	var reqBody io.Reader
	switch strings.ToLower(*method) {
	case "put":
		url = fmt.Sprintf("http://localhost:%d/store/%s", PORT, *key)
		reqBody = strings.NewReader(*value)
	case "get":
		url = fmt.Sprintf("http://localhost:%d/store/%s", PORT, *key)
		reqBody = nil
	case "delete":
		url = fmt.Sprintf("http://localhost:%d/store/%s", PORT, *key)
		reqBody = nil
	default:
		_, err := fmt.Fprintln(os.Stderr, "Error: invalid method, must be 'put', 'get', or 'delete'")
		if err != nil {
			log.Println(err.Error())
			return
		}
		os.Exit(1)
	}

	client := &http.Client{}
	req, err := http.NewRequest(strings.ToUpper(*method), url, reqBody)
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, "Error creating request:", err)
		if err != nil {
			log.Println(err.Error())
			return
		}
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, "Error executing request:", err)
		if err != nil {
			log.Println(err.Error())
			return
		}
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, "Error reading response body:", err)
		if err != nil {
			log.Println(err.Error())
			return
		}
		os.Exit(1)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		fmt.Println(string(body))
	case http.StatusNotFound:
		_, err := fmt.Fprintln(os.Stderr, "Error: key not found")
		if err != nil {
			log.Println(err.Error())
			return
		}
	case http.StatusMethodNotAllowed:
		_, err := fmt.Fprintln(os.Stderr, "Error: method not allowed")
		if err != nil {
			log.Println(err.Error())
			return
		}
	default:
		_, err := fmt.Fprintln(os.Stderr, "Error:", resp.Status)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}

func printUsage() {
	usage := `Usage: client -m=METHOD --key=KEY [--value=VALUE]

METHOD: put, get, or delete
KEY: The key to interact with
VALUE: The value to set (only for put method)

Examples:
  Set a key-value pair:
    client -m=put --key=foo --value=bar

  Get the value of a key:
    client -m=get --key=foo

  Delete a key:
    client -m=delete --key=foo
`

	_, err := fmt.Fprintln(os.Stderr, usage)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
