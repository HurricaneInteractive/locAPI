package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/api/sheets/v4"

	bootstrap "locapi"
)

// LocData is the structure that will be returned by the
// location API
type LocData struct {
	Day     string `json:"day"`
	Time    string `json:"time"`
	Address string `json:"address"`
}

func valuesMapper(resp *sheets.ValueRange) []LocData {
	data := []LocData{}

	for _, row := range resp.Values {
		d := LocData{
			Day:     "",
			Time:    "",
			Address: fmt.Sprintf(`%v`, row[2]),
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
