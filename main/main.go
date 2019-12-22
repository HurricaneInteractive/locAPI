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
	timeRe := regexp.MustCompile(`(\d+:\d+)`)
	yearRe := regexp.MustCompile(`(\d{4})`)
	dayRe := regexp.MustCompile(`(^\w+ \d)`)
	todRe := regexp.MustCompile(`(AM|PM)`)
	time := timeRe.FindAllString(datetime, -1)[0]
	year := yearRe.FindAllString(datetime, -1)[0]
	day := dayRe.FindAllString(datetime, -1)[0]
	tod := todRe.FindAllString(datetime, -1)[0]

	return DateTime{
		Day:  day,
		Year: year,
		Time: time,
		TOD:  tod,
	}
}

func valuesMapper(resp *sheets.ValueRange) []LocData {
	data := []LocData{}

	for _, row := range resp.Values {
		d := LocData{
			DateTime: separateDateTime(fmt.Sprintf(`%v`, row[0])),
			Address:  fmt.Sprintf(`%v`, row[2]),
			State:    fmt.Sprintf(`%v`, row[1]),
		}
		data = append(data, d)
	}

	return data
}

func sheetDump(w http.ResponseWriter, r *http.Request) {
	sheetID := os.Getenv("GS_SHEET_ID")
	sheetName := os.Getenv("GS_SHEET_NAME")
	sheetRange := os.Getenv("GS_SHEET_RANGE")

	client := bootstrap.Sheets()
	resp, err := bootstrap.Values(client, sheetID, sheetName, sheetRange)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		payload := fmt.Sprintf(`{"message": "Unable to retrieve data from sheet", "error": "%v"}`, err)
		w.Write([]byte(payload))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	// payload := fmt.Sprintf(`[{"data": %v}]`, valuesMapper(resp))
	payload, _ := json.Marshal(valuesMapper(resp))
	w.Write([]byte(payload))
}

func logError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %v", msg, err)
	}
}

func main() {
	err := godotenv.Load()
	logError(err, "Error loading .env file")

	r := mux.NewRouter()
	/**
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("", home).Methods(http.MethodGet)
	api.HandleFunc("/user/{userID}/comment/{commentID}", params).Methods(http.MethodGet)
	*/
	r.HandleFunc("/", sheetDump).Methods(http.MethodGet)

	fmt.Printf("Lisiting on port :8080. Visit http://localhost:8080\n")
	log.Fatal(http.ListenAndServe(":8080", r))
}
