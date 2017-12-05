package main

import (
	"net/http"
	"fmt"
	"log"
	"os"
	"io"
	"strconv"
	"net/http/httputil"
)

var (
	HOST          = os.Getenv("HOST")
	PORT          = os.Getenv("PORT")
	IS_BACKEND, _ = strconv.ParseBool(os.Getenv("IS_BACKEND"))
	BACKEND_URL   = os.Getenv("BACKEND_URL")
	HEADERS       = []string{
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-parentspanid",
		"x-b3-flags",
		"x-ot-span-context",
	}
)

func main() {

	if HOST == "" {
		HOST = "0.0.0.0"
	}

	if PORT == "" {
		PORT = "8080"
	}

	if BACKEND_URL == "" {
		BACKEND_URL = "http://0.0.0.0:1337"
	}

	if IS_BACKEND {

		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

			hostname, err := os.Hostname()
			if err != nil {
				fmt.Println("Error: %s", err.Error())
				return
			}

			writer.Write([]byte(hostname))
		})

	} else {

		http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
			_response, err := http.Get(BACKEND_URL)
			if err != nil {
				writer.WriteHeader(500)
				return
			}

			writer.WriteHeader(_response.StatusCode)
		})

		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

			_req, _ := http.NewRequest("GET", BACKEND_URL, nil)

			for _,headerKey := range HEADERS {
					headerValue := request.Header.Get(headerKey)
					if headerValue != "" {
						_req.Header.Set(headerKey,headerValue)
					}
			}

			_response, err := http.Client{}.Do(_req)
			if err != nil {
				fmt.Println(err.Error())
				writer.WriteHeader(500)
				return
			}

			writer.WriteHeader(_response.StatusCode)

			if _response.StatusCode == http.StatusInternalServerError {
				return
			}

			hostname, _ := os.Hostname()

			writer.Write([]byte(fmt.Sprintf("\nFRONTEND: %s", hostname)))
			writer.Write([]byte("\nBACKEND: "))

			_, err = io.Copy(writer, _response.Body)
			if err != nil {
				fmt.Printf("Error copying from backend writer to frontend reader %s\n", err.Error())
				return
			}

			writer.Write([]byte("\n\n"))

			_bytes, err := httputil.DumpRequest(request, true)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			writer.Write(_bytes)
		})

	}

	fmt.Printf("IS_BACKEND: %v\nListening on: %s:%s \n", IS_BACKEND, HOST, PORT)
	if !IS_BACKEND {
		fmt.Printf("Connecting to backend: %s\n", BACKEND_URL)
	}
	err := http.ListenAndServe(HOST+":"+PORT, http.DefaultServeMux)
	if err != nil {
		log.Fatal(err)
	}
}
