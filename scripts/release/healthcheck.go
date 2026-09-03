package main

import (
	"context"
	"net/http"
	"os"
	"time"
)

func main() {
	requestContext, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, "http://127.0.0.1:3000/api/status", nil)
	if err != nil {
		os.Exit(1)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		os.Exit(1)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
