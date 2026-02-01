package main

import (
	"context"
	"example/golang-learn/controllers"
	"example/golang-learn/models"
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

	newCarpark := models.Carpark{
		ID:         123,
		Name:       "SB17",
		PostalCode: "750503",
		Address:    "504 MONTREAL DRIVE MONTREAL SPRING SINGAPORE 750504",
		Location: models.Location{
			Type:        "Point",
			Coordinates: []float64{103.823678171092, 1.45089251057067},
		},
		Lots: []models.Lot{
			{Level: "5A", LotNumber: "355"},
		},
		Vehicles: []models.Vehicle{},
	}
	err = carparkService.InsertCarpark(&newCarpark)

	newVehicle := models.Vehicle{
		ID:             13,
		MakeName:       "Mazda",
		ModelName:      "2",
		PlateNumber:    "SLR9553A",
		Seats:          5,
		Images:         []string{"c57ab461-9df3-43ad-b541-051cc95c8c45_car.png"},
		PriceGroupName: "Standard",
	}
	err = carparkService.AddVehicleToCarpark("SB17", &newVehicle)
	if err != nil {
		fmt.Println(err)
		fmt.Println("error adding vehicle to carpark")
	}

	carpark, err := carparkService.GetCarpark("750503")
	fmt.Println(carpark)

	carparks, err := carparkService.GetCarparksByDistance(103.820052, 1.449466)
	for _, c := range carparks {
		fmt.Printf("Carpark: %v | Distance: %.2f meters\n", c.Name, c.Distance)
	}
	//err = carparkService.DeleteVehicleFromCarpark("SB17", "SLR9553A")

	//err = carparkService.UpdatePostalCode(collection, "SB17", "750503")
	//if err != nil {
	//	fmt.Println("Could not update carpark")
	//	fmt.Println(err)
	//} else {
	//	fmt.Println("Carpark updated")
	//	fmt.Println(carpark)
	//}

	settingService := services.NewSettingService(settingCollection)
	_ = settingService.Set("RadiusKm", "20")
	//_ = settingService.Set("NewKey")
	radius, _ := settingService.GetInt("RadiusKm", 20)
	fmt.Printf("RadiusKm: %v\n", radius)
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
