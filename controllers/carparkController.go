package controllers

import (
	"context"
	"encoding/json"
	"example/golang-learn/dtos"
	"example/golang-learn/models"
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

	results, err := c.carparkService.GetAvailableVehicles(request.Longitude, request.Latitude, request.Start, request.End)
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

	err = c.carparkService.AddScheduleToVehicle(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (c *CarparkController) RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	var request dtos.AddScheduleRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = c.carparkService.DeleteScheduleFromVehicle(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *CarparkController) AddCarpark(w http.ResponseWriter, r *http.Request) {
	var request models.Carpark

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.carparkService.AddCarpark(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *CarparkController) AddVehicle(w http.ResponseWriter, r *http.Request) {
	var request dtos.AddVehicleRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.carparkService.AddVehicleToCarpark(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *CarparkController) RemoveVehicle(w http.ResponseWriter, r *http.Request) {
	var request dtos.RemoveVehicleRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.carparkService.RemoveVehicleFromCarpark(request.CarparkName, request.PlateNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
