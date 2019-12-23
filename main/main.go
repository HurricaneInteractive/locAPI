package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/api/sheets/v4"

	bootstrap "locapi"
)

// DateTime is the breakdown of time entered/exited
type DateTime struct {
	Day  string `json:"day"`
	Year string `json:"year"`
	Time string `json:"time"`
	TOD  string `json:"tod"`
}

// LocData is the structure that will be returned by the
// location API
type LocData struct {
	DateTime DateTime `json:"datetime"`
	Address  string   `json:"address"`
	State    string   `json:"state"`
}

func separateDateTime(datetime string) DateTime {
	re := regexp.MustCompile(`(^\w+ \d+)|(\d{4})|(\d{2}:\d{2})|(AM|PM)`)
	matches := re.FindAllString(datetime, -1)

	return DateTime{
		Day:  matches[0],
		Year: matches[1],
		Time: matches[2],
		TOD:  matches[3],
	}
}

func valuesMapper(resp *sheets.ValueRange) (data []LocData) {
	for _, row := range resp.Values {
		d := LocData{
			DateTime: separateDateTime(fmt.Sprintf(`%v`, row[0])),
			Address:  fmt.Sprintf(`%v`, row[2]),
			State:    fmt.Sprintf(`%v`, row[1]),
		}
		data = append(data, d)
	}

	return
}

func respondWithError(w http.ResponseWriter, status int, msg string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	payload := fmt.Sprintf(`{"message": "%s", "error": "%v"}`, msg, err)
	w.Write([]byte(payload))
}

func sheetDump(w http.ResponseWriter, r *http.Request) {
	sheetID := os.Getenv("GS_SHEET_ID")
	sheetName := os.Getenv("GS_SHEET_NAME")
	sheetRange := os.Getenv("GS_SHEET_RANGE")

	client := bootstrap.Sheets()
	resp, err := bootstrap.Values(client, sheetID, sheetName, sheetRange)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to retrieve data from sheet", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	payload, _ := json.Marshal(valuesMapper(resp))
	w.Write([]byte(payload))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("", sheetDump).Methods(http.MethodGet)

	fmt.Printf("Lisiting on port :8080. Visit http://localhost:8080\n")
	log.Fatal(http.ListenAndServe(":8080", r))
}
