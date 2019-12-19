package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/sheets/v4"

	bootstrap "locapi"
)

func logError(err error, msg string) {
	if err != nil {
		log.Fatal("%s: %v", msg, err)
	}
}

func main() {
	err := godotenv.Load()
	logError(err, "Error loading .env file")

	sheetID := os.Getenv("GS_SHEET_ID")

	if sheetID == "" {
		log.Fatal("No sheet id defined")
	}

	client := bootstrap.Sheets()
	srv, err := sheets.New(client)
	logError(err, "Unable to retrieve Sheets client")

	readRange := "Class Data!A2:E"
	resp, err := srv.Spreadsheets.Values.Get(sheetID, readRange).Do()
	logError(err, "Unable to retrieve data from sheet")

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s\n", row[0], row[1])
		}
	}
}
