package main

import (
	"context"
	"example/golang-learn/controllers"
	"example/golang-learn/services"
	"fmt"
	"net/http"
	"os"
	"sync"

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

	userService := services.NewUserService(ctx)
	userController := controllers.NewUserController(ctx, userService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)

	mux.HandleFunc("POST /users", userController.CreateUser)
	mux.HandleFunc("GET /users/{id}", userController.GetUser)
	mux.HandleFunc("DELETE /users/{id}", userController.DeleteUser)

	fmt.Println("Server listening to :8081")
	http.ListenAndServe(":8081", mux)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}
