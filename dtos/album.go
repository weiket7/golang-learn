package dtos

type Actor struct {
	ActorId string  `json:"id"`
	Title   string  `json:"title"`
	Artist  string  `json:"artist"`
	Price   float64 `json:"price"`
}
