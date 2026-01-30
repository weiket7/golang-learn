package main

import (
	"context"
	"encoding/json"
	"example/golang-learn/helpers/errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	z "github.com/Oudwins/zog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//https://go.dev/doc/tutorial/web-service-gin

type User struct {
	Name string `json:"name"`
}

var userCache = make(map[int]User)

var cacheMutex sync.RWMutex

// read mode, write mode or read and write
// blocks all read and write when mutex is locked
// mutex in general is a safe way to synchronise data in multithreaded app

type Config struct {
	Env                string
	DbConnectionString string
}

func main() {
	//ctx := context.Background()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logger := zerolog.New(os.Stdout)
	ctx := context.Background()

	ctx = logger.WithContext(ctx)

	//set min level info
	//zerolog.SetGlobalLevel(zerolog.InfoLevel)

	//appEnv := os.Getenv("APP_ENV")
	//fmt.Println("Application Environment:", appEnv)

	v := viper.New()
	v.SetConfigFile(".env")
	err := v.ReadInConfig()
	if err != nil {
		//log.Fatal().Err(err).Msg("Error reading config file")
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	env := v.GetString("ENVIRONMENT")
	fmt.Println("environment: ", env)

	cfg := &Config{}
	err = v.Unmarshal(&cfg)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
		return
	}

	log.Log().Interface("config", cfg).Msg("config loaded")

	log.Print("hello world")
	log.Log().
		Str("foo", "bar").
		Msg("")

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

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Log().Interface("user", user).Msg("user")

	var userSchema = z.Struct(z.Shape{
		"name": z.String().Required().Min(3).Max(10),
		//"age":  z.Int().GT(18),
	})
	errs := userSchema.Validate(&user)
	if errs != nil {
		fmt.Println(errs)

		errorResponse := errors.NewValidationError(errs)
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	//if user.Name == "" {
	//	http.Error(w, "Name is required", http.StatusBadRequest)
	//	return
	//}

	cacheMutex.Lock()
	userId := len(userCache) + 1
	userCache[userId] = user
	cacheMutex.Unlock()
	fmt.Println("create user id:", userId)

	w.WriteHeader(http.StatusCreated)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}
