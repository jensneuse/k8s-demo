package main

import (
	"net/http"
	"fmt"
	"log"
	"os"
	"strconv"
	"net/http/httputil"
	"io/ioutil"
	"strings"
	"time"
)

var (
	HOST            = os.Getenv("HOST")
	PORT            = os.Getenv("PORT")
	NAME            = os.Getenv("NAME")
	IS_BACKEND, _   = strconv.ParseBool(os.Getenv("IS_BACKEND"))
	BACKEND_URL     = os.Getenv("BACKEND_URL")
	BACKEND_LATENCY = os.Getenv("BACKEND_LATENCY")
	HEADERS         = []string{
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-parentspanid",
		"x-b3-flags",
		"x-ot-span-context",
	}
	client = http.Client{}
)

func main() {

	client.Timeout = time.Second * 5

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

			latency,err := strconv.Atoi(BACKEND_LATENCY)
			if err == nil {
				time.Sleep(time.Second * time.Duration(latency))
			}

			hostname, err := os.Hostname()
			if err != nil {
				fmt.Println("Error: %s", err.Error())
				return
			}

			if NAME != "" {
				writer.Write([]byte(fmt.Sprintf("Servicename: %s - Hostname: %s\n", NAME, hostname)))
			} else {
				writer.Write([]byte(hostname))
			}
		})

	} else {

		http.HandleFunc("/favicon.ico", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
			_response, err := http.Get(BACKEND_URL)
			if err != nil {
				writer.WriteHeader(500)
				return
			}

			writer.WriteHeader(_response.StatusCode)
		})

		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

			urls := strings.Split(BACKEND_URL, ",")
			responses := fanIn(multiGenerateBlocking(request, urls...)...)
			//responses := multiGenerateSerial(request, urls...)

			hostname, _ := os.Hostname()

			if NAME != "" {
				writer.Write([]byte(fmt.Sprintf("\nFRONTEND - Servicename: %s %s",NAME, hostname)))
			} else {
				writer.Write([]byte(fmt.Sprintf("\nFRONTEND: %s", hostname)))
			}

			for i := 0; i < len(urls); i++ {
				response := <-responses
				writer.Write([]byte(fmt.Sprintf("\nBACKEND - %s - %d", response.body, response.statusCode)))
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

type blockingResponse struct {
	statusCode int
	body       string
}

func multiGenerateSerial(request *http.Request, urls ...string) (<-chan blockingResponse) {

	c := make(chan blockingResponse)

	go func() {
		for _, url := range urls {
			getBlocking(url, request, c)
		}
	}()

	return c
}

func multiGenerateBlocking(request *http.Request, urls ...string) ([]<-chan blockingResponse) {

	var responseChannels []<-chan blockingResponse

	for _, url := range urls {
		responseChannels = append(responseChannels, generateGetBlocking(url, request))
	}

	return responseChannels
}

func generateGetBlocking(url string, request *http.Request) (chan blockingResponse) {
	c := make(chan blockingResponse)
	go getBlocking(url, request, c)
	return c
}

func getBlocking(url string, request *http.Request, c chan blockingResponse) {
	_req, _ := http.NewRequest("GET", url, nil)

	for _, headerKey := range HEADERS {
		headerValue := request.Header.Get(headerKey)
		if headerValue != "" {
			_req.Header.Set(headerKey, headerValue)
		}
	}

	cookies := request.Cookies()

	for _, cookie := range cookies {
		_req.AddCookie(cookie)
	}

	response, err := client.Do(_req)
	if err != nil {
		fmt.Println(err.Error())
		c <- blockingResponse{
			statusCode: http.StatusInternalServerError,
			body:       "",
		}
		return
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
		c <- blockingResponse{
			statusCode: http.StatusInternalServerError,
			body:       "",
		}
		return
	}

	c <- blockingResponse{
		statusCode: response.StatusCode,
		body:       string(responseBody),
	}
}

func fanIn(inputs ...<-chan blockingResponse) (chan blockingResponse) {

	c := make(chan blockingResponse)

	for _, input := range inputs {
		go func(inputChannel <-chan blockingResponse) {
			for {
				c <- <-inputChannel
			}
		}(input)
	}

	return c
}
