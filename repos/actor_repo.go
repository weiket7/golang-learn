package repos

import (
	"example/golang-gin/dtos"

	"github.com/jmoiron/sqlx"
)

func all() []dtos.Actor {
	var actors []dtos.Actor
	actors = append(actors, dtos.Actor{Artist: "Hello"})
	return actors
}

type ActorRepo struct {
	db *sqlx.DB
}

func NewActorRepo(db *sqlx.DB) *ActorRepo {
	return &ActorRepo{db: db}
}
