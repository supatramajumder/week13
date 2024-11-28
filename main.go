package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type TimeResponse struct {
	CurrentTime string `json:"current_time"`
}

var db *sql.DB

func main() {
	var err error

	db, err = sql.Open("mysql", "root:Password123@tcp(127.0.0.1:3306)/SUPATRA_WEEK13")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/current-time", getCurrentTime).Methods("GET")
	r.HandleFunc("/log-times", getLoggedTimes).Methods("GET")

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getCurrentTime(w http.ResponseWriter, r *http.Request) {
	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		http.Error(w, "Failed to load timezone", http.StatusInternalServerError)
		log.Println("Error loading timezone:", err)
		return
	}

	currentTime := time.Now().In(loc)
	formattedTime := currentTime.Format("2006-01-02 15:04:05") // BUG FIX to insert records in toronto time zone

	_, err = db.Exec("INSERT INTO supatra_toronto_time (timestamp) VALUES (?)", formattedTime)
	if err != nil {
		http.Error(w, "Failed to log time", http.StatusInternalServerError)
		log.Println("Error logging time:", err)
		return
	}

	response := TimeResponse{CurrentTime: currentTime.Format(time.RFC3339)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getLoggedTimes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, timestamp FROM time_log")
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		log.Println("Error fetching logs:", err)
		return
	}
	defer rows.Close()

	var logs []TimeResponse
	for rows.Next() {
		var id int
		var timestamp time.Time
		if err := rows.Scan(&id, &timestamp); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		logs = append(logs, TimeResponse{CurrentTime: timestamp.Format(time.RFC3339)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
