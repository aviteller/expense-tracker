package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

type Transaction struct {
	ID     int     `json:"id"`
	Text   string  `json:"text"`
	Amount float32 `json:"amount"`
}

func main() {
	router := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedHeaders: []string{"*"},
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"},
	})

	initTables()

	router.HandleFunc("/transaction", getTransactions).Methods("GET")
	router.HandleFunc("/transaction/{id}", deleteTransaction).Methods("DELETE")
	router.HandleFunc("/transaction", addTransaction).Methods("POST")

	router.PathPrefix("/").Handler(http.FileServer(rice.MustFindBox("static/public").HTTPBox()))

	port := "9000"
	if port == "" {
		port = "8000"
	}

	open("http://localhost:9000/")
	err := http.ListenAndServe(":"+port, c.Handler(router))
	if err != nil {
		fmt.Println(err)
	}
}

func getDB() *sql.DB {
	database, err := sql.Open("sqlite3", "./transactions.db")
	if err != nil {
		fmt.Println(err)
	}
	return database
}

var deleteTransaction = func(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]

	database := getDB()
	stmt, err := database.Prepare("update `transactions` SET `deleted` = 1, `updated_at` = ? where id=?")
	if err != nil {
		fmt.Println(err)
	}
	t := time.Now()
	stmt.Exec(t.Format("2006-01-02T15:04:05Z07:00"), id)

	res := message(true, "success")
	res["data"] = id
	respond(w, res)
}

var addTransaction = func(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	err := json.NewDecoder(r.Body).Decode(&t)

	if err != nil {
		fmt.Println(err)
		respond(w, message(false, "Invalid request"))
		return
	}

	database := getDB()
	statement, err := database.Prepare("INSERT INTO `transactions` (`text`,`amount`,`created_at`,`updated_at`) VALUES (?,?,?,?)")

	if err != nil {
		fmt.Println(err)
	}

	time := time.Now()

	result, err := statement.Exec(t.Text, t.Amount, time.Format("2006-01-02T15:04:05Z07:00"), time.Format("2006-01-02T15:04:05Z07:00"))
	if err != nil {
		fmt.Println(err)
	}

	lastid, _ := result.LastInsertId()

	t.ID = int(lastid)

	res := message(true, "success")
	res["transaction"] = t
	respond(w, res)

}

var getTransactions = func(w http.ResponseWriter, r *http.Request) {
	var transactions []Transaction
	database := getDB()
	rows, err := database.Query("SELECT id, text, amount FROM `transactions` WHERE deleted = 0")
	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.ID, &t.Text, &t.Amount)
		if err != nil {
			fmt.Println(err)
		}
		transactions = append(transactions, t)

	}

	res := message(true, "success")

	if len(transactions) > 0 {

		res["data"] = transactions
	}

	respond(w, res)

}

func initTables() {
	getDB().Exec("CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY autoincrement, text text, amount double,  created_at text, updated_at text, deleted int DEFAULT 0)")
}

func respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)
}

func message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
