package services

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SettingService struct {
	coll *mongo.Collection
}

func NewSettingService(coll *mongo.Collection) *SettingService {
	return &SettingService{
		coll: coll,
	}
}

func (s *SettingService) GetString(key string, defaultVal string) (string, error) {
	var result map[string]any
	err := s.coll.FindOne(context.Background(), bson.M{}).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("GetString err: %w", err)
	}

	return result[key].(string), nil
}

func (s *SettingService) GetInt(key string, defaultVal int) (int, error) {
	var result map[string]any
	err := s.coll.FindOne(context.Background(), bson.M{}).Decode(&result)
	if err != nil {
		return defaultVal, fmt.Errorf("database error: %w", err)
	}

	val, exists := result[key]
	if !exists {
		return defaultVal, nil
	}

	// try to assert as int64 (MongoDB default for numbers)
	if i64, ok := val.(int64); ok {
		return int(i64), nil
	}

	if i, ok := val.(int); ok {
		return i, nil
	}

	return defaultVal, fmt.Errorf("key '%s' is type %T, not an integer", key, val)
}

func (s *SettingService) Set(key string, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{} //empty filter bson.M{} because treat this collection as a single-document "Singleton".

	update := bson.M{
		"$set": bson.M{
			key:         value,
			"updatedAt": time.Now(),
		},
	}

	// Upsert: true ensures the doc is created if the collection is empty
	opts := options.UpdateOne().SetUpsert(true)

	_, err := s.coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update mongo setting: %w", err)
	}

	return nil
}
