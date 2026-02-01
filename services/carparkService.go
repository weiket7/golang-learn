package services

import (
	"context"
	"example/golang-learn/models"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CarparkService struct {
	coll *mongo.Collection
}

func NewCarparkService(coll *mongo.Collection) *CarparkService {
	return &CarparkService{
		coll: coll,
	}
}

func (s *CarparkService) InsertCarpark(newCarpark *models.Carpark) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.coll.InsertOne(ctx, newCarpark)
	if err != nil || result == nil {
		return err
	}

	//fmt.Printf("Inserted Carpark with ID: %v\n", result.InsertedID)
	return nil
}

func (s *CarparkService) UpdatePostalCode(carparkName string, postalCode string) error {
	filter := bson.M{"name": carparkName}

	update := bson.M{
		"$set": bson.M{
			"postalCode": postalCode,
		},
	}

	result, err := s.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil || result == nil {
		return fmt.Errorf("UpdatePostalCode could not update carpark %s", carparkName)
	}

	//fmt.Printf("Matched %v documents and updated %v documents.\n", result.MatchedCount, result.ModifiedCount)
	return nil
}

func (s *CarparkService) GetCarpark(postalCode string) (*models.Carpark, error) {
	filter := bson.M{"postalCode": postalCode}

	var carpark models.Carpark
	err := s.coll.FindOne(context.TODO(), filter).Decode(&carpark)
	if err != nil {
		return nil, fmt.Errorf("GetCarpark carpark %s not found", postalCode)
	}
	return &carpark, nil
}

func (s *CarparkService) GetCarparksByDistance(lon float64, lat float64) ([]models.Carpark, error) {
	pipeline := mongo.Pipeline{
		{
			{Key: "$geoNear", Value: bson.D{
				{Key: "near", Value: bson.D{
					{Key: "type", Value: "Point"},
					{Key: "coordinates", Value: []float64{lon, lat}},
				}},
				{Key: "distanceField", Value: "dist"}, // Field added to output
				{Key: "spherical", Value: true},       // Required for 2dsphere
			}},
		},
		// Optional: Only return the car park name and the distance
		//{
		//	{Key: "$project", Value: bson.D{
		//		{Key: "name", Value: 1},
		//		{Key: "dist", Value: 1},
		//		{Key: "_id", Value: 0},
		//	}},
		//},
	}

	cursor, err := s.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("GetCarparksByDistance could not aggregate %w", err)
	}

	var results []models.Carpark
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, fmt.Errorf("GetCarparksByDistance could not parse %w", err)
	}

	return results, nil
}

func (s *CarparkService) DeleteVehicleFromCarpark(carparkName string, plateNumber string) error {
	// 1. Filter: Find the specific carpark
	filter := bson.M{"name": carparkName}

	// 2. Update: $pull from the 'vehicles' array where '_id' matches
	update := bson.M{
		"$pull": bson.M{
			"vehicles": bson.M{"plateNumber": plateNumber},
		},
	}

	// 3. Execute
	result, err := s.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil || result == nil || result.ModifiedCount == 0 {
		return fmt.Errorf("DeleteVehicleFromCarpark could not delete %s %w", carparkName, err)
	}

	return nil
}

func (s *CarparkService) AddVehicleToCarpark(carparkName string, vehicle *models.Vehicle) error {
	// 2. Filter: Find the specific carpark
	filter := bson.M{"name": carparkName}

	// 3. Update: $push the vehicle into the 'vehicles' slice
	update := bson.M{
		"$push": bson.M{
			"vehicles": vehicle,
		},
	}

	// 4. Execute
	result, err := s.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	if result == nil || result.ModifiedCount == 0 {
		return fmt.Errorf("AddVehicleToCarpark could not add vehicle %s to carpark %s", vehicle.ID, carparkName)
	}
	return nil
}
