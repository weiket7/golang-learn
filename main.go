package main

import (
	"context"
	"example/golang-learn/controllers"
	"example/golang-learn/services"
	"example/golang-learn/utilities/db"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Env                string
	DbConnectionString string
}

func main() {
	fmt.Println("starting service")

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout)

	//set min level info
	//zerolog.SetGlobalLevel(zerolog.InfoLevel)

	ctx := context.Background()
	ctx = logger.WithContext(ctx)

	//appEnv := os.Getenv("APP_ENV")
	//fmt.Println("Application Environment:", appEnv)

	v := viper.New()
	v.SetConfigFile(".env")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("could not read config: %w", err))
	}

	cfg := &Config{}
	err = v.Unmarshal(&cfg)
	if err != nil {
		panic(fmt.Errorf("could not parse config: %w", err))
	}

	log.Log().Interface("config", cfg).Msg("config loaded")

	client, err := db.Connect(v.GetString("DB_CONNECTION_STRING"))
	if err != nil {
		panic(fmt.Errorf("could not connect to mongo: %w", err))
	}

	//defer client.Disconnect(ctx)
	defer func() {
		//Immediately Invoked Function Expression (IIFE) so can write logic like an if statement inside defer
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	database := client.Database("getgo")
	collection := database.Collection("carparks")
	settingCollection := database.Collection("settings")

	env := v.GetString("ENVIRONMENT")
	fmt.Println("started environment: ", env)

	//log.Print("hello world")
	//log.Log().
	//	Str("foo", "bar").
	//	Msg("")

	carparkService := services.NewCarparkService(collection)
	settingService := services.NewSettingService(settingCollection)
	_ = settingService.Set("RadiusKm", "20")
	radius, _ := settingService.GetInt("RadiusKm", 20)
	fmt.Printf("RadiusKm: %v\n", radius)

	userService := services.NewUserService(ctx)
	userController := controllers.NewUserController(ctx, userService)
	carparkController := controllers.NewCarparkController(ctx, carparkService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)

	mux.HandleFunc("POST /users", userController.CreateUser)
	mux.HandleFunc("GET /users/{id}", userController.GetUser)
	mux.HandleFunc("DELETE /users/{id}", userController.DeleteUser)

	mux.HandleFunc("GET /carparks", carparkController.GetCarparks)
	mux.HandleFunc("POST /carparks", carparkController.AddCarpark)
	mux.HandleFunc("POST /vehicles", carparkController.AddVehicle)
	mux.HandleFunc("DELETE /vehicles", carparkController.RemoveVehicle)
	mux.HandleFunc("POST /schedules", carparkController.AddSchedule)
	mux.HandleFunc("DELETE /schedules", carparkController.RemoveSchedule)

	fmt.Println("Server listening to :8081")
	http.ListenAndServe(":8081", mux)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}
