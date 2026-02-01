package dtos

import "time"

type CarparksRequest struct {
	Latitude       float64   `json:"latitude" validate:"required,gte=-90,lte=90"`
	Longitude      float64   `json:"longitude" validate:"required,gte=-180,lte=180"`
	PriceGroupIds  []int     `json:"priceGroupIds"`
	VehicleTypeIds []int     `json:"vehicleTypeIds"`
	NumSeats       int       `json:"numSeats"`
	Start          time.Time `json:"start" validate:"required"`
	End            time.Time `json:"end" validate:"required"`
}
