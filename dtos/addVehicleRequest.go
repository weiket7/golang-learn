package dtos

import "example/golang-learn/models"

type AddVehicleRequest struct {
	CarparkName    string       `json:"carparkName"`
	MakeName       string       `json:"makeName"`
	ModelName      string       `json:"modelName"`
	PlateNumber    string       `json:"plateNumber"`
	Seats          int          `json:"seats"`
	PriceGroupName string       `json:"priceGroupName"`
	PriceGroupId   int          `json:"priceGroupId"`
	Images         []string     `json:"images"`
	Lots           []models.Lot `json:"lots"`
}

type RemoveVehicleRequest struct {
	CarparkName string `json:"carparkName"`
	PlateNumber string `json:"plateNumber"`
}
