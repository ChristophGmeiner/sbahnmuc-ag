package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	clientID := os.Getenv("db_client_id")
	clientSecret := os.Getenv("db_client_secret_api_key")
	if clientID == "" {
		fmt.Println("No env")
		return
	}

	urls := []string{
		"https://apis.deutschebahn.com/db-api-marketplace/apis/timetables/v1/station/München",
		"https://apis.deutschebahn.com/db-api-marketplace/apis/station-data/v2/stations?searchstring=*M%C3%BCnchen*",
	}

	for _, url := range urls {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("DB-Client-Id", clientID)
		req.Header.Set("DB-Api-Key", clientSecret)
		req.Header.Set("Accept", "application/xml")
		
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("URL: %s\nStatus: %d\nBody (first 200 chars): %s\n\n", url, resp.StatusCode, string(body)[:min(200, len(body))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
