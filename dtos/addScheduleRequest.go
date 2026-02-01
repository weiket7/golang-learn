package dtos

type AddScheduleRequest struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	CarparkId int    `json:"carparkId"`
	VehicleId int    `json:"vehicleId"`
	BookingId int    `json:"bookingId"`
}
