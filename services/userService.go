package services

import (
	"context"
	"errors"
	"example/golang-learn/dtos"
	"sync"
)

type UserService struct {
	ctx        context.Context
	cacheMutex sync.RWMutex
	userCache  map[int]dtos.User
	//log log.Utility
}

func NewUserService(ctx context.Context) *UserService {
	return &UserService{
		ctx:       ctx,
		userCache: make(map[int]dtos.User),
	}
}

func (s *UserService) CreateUser(user dtos.User) int {
	s.cacheMutex.Lock()
	userId := len(s.userCache) + 1
	s.userCache[userId] = user
	s.cacheMutex.Unlock()
	return userId
}

func (s *UserService) DeleteUser(id int) error {
	if _, ok := s.userCache[id]; !ok {
		return errors.New("user not found")
	}

	s.cacheMutex.Lock()
	delete(s.userCache, id)
	s.cacheMutex.Unlock()

	return nil
}

func (s *UserService) GetUser(id int) (*dtos.User, error) {
	s.cacheMutex.RLock()
	user, ok := s.userCache[id]
	s.cacheMutex.RUnlock()

	if !ok {
		return nil, errors.New("user not found")
	}

	return &user, nil
}
