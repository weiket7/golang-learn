package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

//https://go.dev/doc/tutorial/web-service-gin

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

var schema = `
CREATE TABLE person (
    first_name text,
    last_name text,
    email text
);

CREATE TABLE place (
    country text,
    city text NULL,
    telcode integer
)`

type Place struct {
	Country string
	City    sql.NullString
	TelCode int
}

type User struct {
	Name string `json: "name"`
}

var userCache = make(map[int]User)

var cacheMutex sync.RWMutex

// read mode, write mode or read and write
// blocks all read and write when mutex is locked
// mutex in general is a safe way to synchronise data in multithreaded app

func main() {
	//ctx := context.Background()

	//log.Print("hello world")
	//
	//db, err := sqlx.Connect("mysql", "root:GetGoing!@tcp(127.0.0.1:3306)/golang?multiStatements=true")
	//if err != nil {
	//	log.Fatal().Err(err).Msg("failed to connect to database")
	//}
	//
	////db.MustExec(schema)
	//
	////tx := db.MustBegin()
	////tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "Jason", "Moiron", "jmoiron@jmoiron.net")
	////tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES (?, ?, ?)", "John", "Doe", "johndoeDNE@gmail.net")
	////tx.MustExec("INSERT INTO place (country, city, telcode) VALUES (?, ?, ?)", "United States", "New York", "1")
	////tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Hong Kong", "852")
	////tx.MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Singapore", "65")
	////// Named queries can use structs, so if you have an existing struct (i.e. person := &Person{}) that you have populated, you can pass it in as &person
	////tx.NamedExec("INSERT INTO person (first_name, last_name, email) VALUES (:first_name, :last_name, :email)", &Person{"Jane", "Citizen", "jane.citzen@example.com"})
	////tx.Commit()
	//
	//// Query the database, storing results in a []Person (wrapped in []interface{})
	//var people []dtos.Person
	////people := []dtos.Person{}
	//db.Select(&people, "SELECT * FROM person ORDER BY first_name ASC")
	//jason, john := people[0], people[1]
	//
	//fmt.Printf("%#v\n%#v", jason, john)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)

	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUser)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("Server listening to :8081")
	http.ListenAndServe(":8081", mux)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, ok := userCache[id]; !ok {
		http.NotFound(w, r)
	}

	cacheMutex.Lock()
	delete(userCache, id)
	cacheMutex.Unlock()

	fmt.Println("delete user id:", id)
	w.WriteHeader(http.StatusNoContent)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	fmt.Println("get user id:", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheMutex.RLock()
	user, ok := userCache[id]
	cacheMutex.RUnlock()

	if !ok {
		http.NotFound(w, r)
	}

	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func createUser(writer http.ResponseWriter, request *http.Request) {
	var user User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(writer, "Name is required", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	userId := len(userCache) + 1
	userCache[userId] = user
	cacheMutex.Unlock()
	fmt.Println("create user id:", userId)

	writer.WriteHeader(http.StatusCreated)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}
