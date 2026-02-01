package models

import "time"

type Location struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

type Schedule struct {
	Start time.Time `bson:"start"`
	End   time.Time `bson:"end"`
	Type  string    `bson:"type"`
	Id    int       `bson:"sourceId"`
}

type Vehicle struct {
	Id             int        `bson:"_id"`
	MakeName       string     `bson:"makeName"`
	ModelName      string     `bson:"modelName"`
	PlateNumber    string     `bson:"plateNumber"`
	Seats          int        `bson:"seats"`
	PriceGroupName string     `bson:"priceGroupName"`
	Images         []string   `bson:"images"`
	Schedules      []Schedule `bson:"schedules"`
}

type Carpark struct {
	Id                int       `bson:"_id,omitempty"`
	Name              string    `bson:"name"`
	PostalCode        string    `bson:"postalCode"`
	Location          Location  `bson:"location"`
	Vehicles          []Vehicle `bson:"vehicles"`
	HasSlashedVehicle bool      `bson:"hasSlashedVehicle"`
	Address           string    `bson:"address"`
	Lots              []Lot     `bson:"lots"`
	Distance          float64   `bson:"dist" json:"distance"`
	AvailableVehicles int       `bson:"availableVehicles"` // Matches the added field
}

type Lot struct {
	Level     string `bson:"level"`
	LotNumber string `bson:"lotNumber"`
}
