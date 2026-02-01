package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"example/golang-learn/dtos"
	"example/golang-learn/models"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CarparkService struct {
	coll *mongo.Collection
}

func NewCarparkService(coll *mongo.Collection) *CarparkService {
	return &CarparkService{
		coll: coll,
	}
}

func (s *CarparkService) AddCarpark(newCarpark *models.Carpark) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Find the highest INT id
	// We add a filter to ONLY look at documents where _id is a number (Type 1 or 16/18 in BSON)
	// This prevents the ObjectID error.
	filter := bson.M{"_id": bson.M{"$type": "number"}}
	opts := options.FindOne().SetSort(bson.M{"_id": -1})

	var lastDoc struct {
		ID int `bson:"_id"`
	}

	err := s.coll.FindOne(ctx, filter, opts).Decode(&lastDoc)

	newID := 1
	if err == nil {
		newID = lastDoc.ID + 1
	} else if err != mongo.ErrNoDocuments {
		// If it still fails, it means even the "number" couldn't decode
		return fmt.Errorf("failed to determine next ID: %w", err)
	}

	// 2. Assign and Insert
	newCarpark.Id = newID
	_, err = s.coll.InsertOne(ctx, newCarpark)
	return err
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

func (s *CarparkService) GetAvailableVehicles(lon, lat float64, start, end time.Time) ([]models.Carpark, error) {
	// 1. Ensure UTC for MongoDB compatibility
	start = start.UTC()
	end = end.UTC()

	pipeline := mongo.Pipeline{
		// Stage 1: Geospatial search (20km radius)
		{{Key: "$geoNear", Value: bson.D{
			{Key: "near", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: []float64{lon, lat}},
			}},
			{Key: "distanceField", Value: "dist"},
			{Key: "spherical", Value: true},
			{Key: "maxDistance", Value: 20000},
		}}},

		// Stage 2: Calculate availability
		{{Key: "$addFields", Value: bson.D{
			{Key: "availableVehicles", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$filter", Value: bson.D{
						// Handle potential null vehicles array
						{Key: "input", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$vehicles", bson.A{}}}}},
						{Key: "as", Value: "v"},
						{Key: "cond", Value: bson.D{
							{Key: "$eq", Value: bson.A{
								bson.D{{Key: "$size", Value: bson.D{
									{Key: "$filter", Value: bson.D{
										// Note: Your data uses "schedules" (plural)
										{Key: "input", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$$v.schedules", bson.A{}}}}},
										{Key: "as", Value: "sch"},
										{Key: "cond", Value: bson.D{
											{Key: "$and", Value: bson.A{
												// Overlap: existing.start < requested.end AND existing.end > requested.start
												bson.D{{Key: "$lt", Value: bson.A{"$$sch.start", end}}},
												bson.D{{Key: "$gt", Value: bson.A{"$$sch.end", start}}},
											}},
										}},
									}},
								}}},
								0, // No overlapping schedules means vehicle is available
							}},
						}},
					}},
				}}},
			}},
		}}}

	cursor, err := s.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}
	defer cursor.Close(context.TODO())

	var results []models.Carpark
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}

	return results, nil
}

func (s *CarparkService) RemoveVehicleFromCarpark(carparkName string, plateNumber string) error {
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

func (s *CarparkService) AddVehicleToCarpark(req *dtos.AddVehicleRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Find the largest Vehicle ID across ALL carparks
	//Since vehicles are inside an array in every carpark document, we use $unwind to flatten them all into one list
	//then $max to find the highest _id currently in the database.
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$vehicles"}}, //
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "maxId", Value: bson.D{{Key: "$max", Value: "$vehicles._id"}}},
		}}},
	}

	cursor, err := s.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return err
	}

	newVehicleId := 1
	if len(results) > 0 && results[0]["maxId"] != nil {
		// Convert the result to int (Mongo usually returns int32 or int64)
		if maxId, ok := results[0]["maxId"].(int32); ok {
			newVehicleId = int(maxId) + 1
		} else if maxId, ok := results[0]["maxId"].(int64); ok {
			newVehicleId = int(maxId) + 1
		}
	}

	// 2. Build the Vehicle object
	vehicle := models.Vehicle{
		Id:          newVehicleId, // Assign the incremented ID
		MakeName:    req.MakeName,
		ModelName:   req.ModelName,
		PlateNumber: req.PlateNumber,
		Seats:       req.Seats,
		Lots:        req.Lots,
		Images:      req.Images,
		Schedules:   []models.Schedule{},
	}

	// 3. Update the specific carpark
	filter := bson.M{"name": req.CarparkName}
	update := bson.M{
		"$push": bson.M{
			"vehicles": vehicle,
		},
	}

	result, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("carpark '%s' not found", req.CarparkName)
	}

	return nil
}

func (s *CarparkService) ImportVehicles(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		vId, _ := strconv.Atoi(record[0])
		numSeats, _ := strconv.Atoi(record[2])
		priceGroupId, _ := strconv.Atoi(record[13])
		cpId, _ := strconv.Atoi(record[4])
		lat, _ := strconv.ParseFloat(record[8], 64)
		lon, _ := strconv.ParseFloat(record[9], 64)

		var cpLots []models.Lot
		json.Unmarshal([]byte(record[15]), &cpLots)

		newVehicle := models.Vehicle{
			Id:           vId,
			MakeName:     record[10],
			ModelName:    record[11],
			PlateNumber:  record[1],
			Seats:        numSeats,
			PriceGroupId: priceGroupId,
			Images:       []string{record[14]},
			Schedules:    []models.Schedule{},
			Lots:         cpLots,
		}

		filter := bson.M{"_id": cpId}

		// atomic update
		update := bson.M{
			// $setOnInsert only runs when the document is being CREATED
			"$setOnInsert": bson.M{
				"name":       record[5],
				"postalCode": record[7],
				"address":    record[6],
				"location": bson.M{
					"type":        "Point",
					"coordinates": []float64{lon, lat},
				},
			},
			// $addToSet adds the vehicle only if it doesn't already exist in the array
			"$addToSet": bson.M{
				"vehicles": newVehicle,
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		_, err = s.coll.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			log.Printf("Failed to upsert carpark %d: %v", cpId, err)
		}
	}

	return nil
}

func (s *CarparkService) AddScheduleToVehicle(req dtos.AddScheduleRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Parse strings to time.Time (MongoDB needs Date objects, not strings)
	startTime, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		return fmt.Errorf("invalid start time: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		return fmt.Errorf("invalid end time: %w", err)
	}

	// 2. Define the filter (Find the Carpark)
	filter := bson.M{"_id": req.CarparkId}

	// 3. Define the Update logic
	// We use "vehicles.$[v].schedules" where [v] is a placeholder for the matched vehicle
	update := bson.M{
		"$push": bson.M{
			"vehicles.$[v].schedules": bson.M{
				"bookingId": req.BookingId,
				"start":     startTime,
				"end":       endTime,
			},
		},
	}

	// 4. Define the ArrayFilter to identify which vehicle in the array gets the update
	opts := options.UpdateOne().SetArrayFilters([]any{
		bson.M{"v._id": req.VehicleId},
	})

	// 5. Execute
	result, err := s.coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("carpark %d or vehicle %d not found", req.CarparkId, req.VehicleId)
	}

	return nil
}

func (s *CarparkService) DeleteScheduleFromVehicle(req dtos.AddScheduleRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Filter the parent Carpark document
	filter := bson.M{"_id": req.CarparkId}

	// 2. Define the Update logic
	// $pull removes the element from the schedules array that matches the bookingId
	update := bson.M{
		"$pull": bson.M{
			"vehicles.$[v].schedules": bson.M{
				"bookingId": req.BookingId,
			},
		},
	}

	// 3. ArrayFilter to identify the specific vehicle inside the array
	opts := options.UpdateOne().SetArrayFilters([]any{
		bson.M{"v._id": req.VehicleId},
	})

	// 4. Execute the update
	result, err := s.coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	// 5. Verify if something was actually found
	if result.MatchedCount == 0 {
		return fmt.Errorf("carpark %d or vehicle %d not found", req.CarparkId, req.VehicleId)
	}

	return nil
}
