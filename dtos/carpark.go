package dtos

type Location struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

type Vehicle struct {
	ID             int      `bson:"_id"`
	MakeName       string   `bson:"makeName"`
	ModelName      string   `bson:"modelName"`
	PlateNumber    string   `bson:"plateNumber"`
	Seats          int      `bson:"seats"`
	PriceGroupName string   `bson:"priceGroupName"`
	Images         []string `bson:"images"`
	// Add other fields as needed
}

type Carpark struct {
	ID                int       `bson:"_id,omitempty"`
	Name              string    `bson:"name"`
	PostalCode        string    `bson:"postalCode"`
	Location          Location  `bson:"location"`
	Vehicles          []Vehicle `bson:"vehicles"`
	HasSlashedVehicle bool      `bson:"hasSlashedVehicle"`
	Address           string    `bson:"address"`
	Lots              []Lot     `bson:"lots"`
	Distance          float64   `bson:"dist" json:"distance"`
}

type Lot struct {
	Level     string `bson:"level"`
	LotNumber string `bson:"lotNumber"`
}
