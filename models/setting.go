package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Setting struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Radius int           `bson:"radius" json:"radius"`
}
