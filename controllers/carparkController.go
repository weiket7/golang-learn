package controllers

import (
	"context"
	"encoding/json"
	"example/golang-learn/dtos"
	"example/golang-learn/services"
	"net/http"
)

type CarparkController struct {
	ctx            context.Context
	carparkService *services.CarparkService
}

func NewCarparkController(ctx context.Context, carparkService *services.CarparkService) *CarparkController {
	return &CarparkController{
		ctx:            ctx,
		carparkService: carparkService,
	}
}

func (c *CarparkController) GetCarparks(w http.ResponseWriter, r *http.Request) {
	var request dtos.CarparksRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results, err := c.carparkService.GetAvailableCarparks(request.Longitude, request.Latitude, request.Start, request.End)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (c *CarparkController) AddSchedule(w http.ResponseWriter, r *http.Request) {
	var request dtos.AddScheduleRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = c.carparkService.AddVehicleSchedule(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
