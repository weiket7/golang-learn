package controllers

import (
	"context"
	"encoding/json"
	"example/golang-learn/dtos"
	"example/golang-learn/helpers/errors"
	"example/golang-learn/services"
	"fmt"
	"net/http"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog/log"
)

type UserController struct {
	ctx         context.Context
	userService *services.UserService
}

func NewUserController(ctx context.Context, service *services.UserService) *UserController {
	return &UserController{
		ctx:         ctx,
		userService: service,
	}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user dtos.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Log().Interface("user", user).Msg("user")

	var userSchema = z.Struct(z.Shape{
		"name": z.String().Required().Min(3).Max(10),
		//"age":  z.Int().GT(18),
	})
	errs := userSchema.Validate(&user)
	if errs != nil {
		fmt.Println(errs)

		errorResponse := errors.NewValidationError(errs)
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	//if user.Name == "" {
	//	http.Error(w, "Name is required", http.StatusBadRequest)
	//	return
	//}

	userId := c.userService.CreateUser(user)
	fmt.Println("User created ", userId)

	w.WriteHeader(http.StatusCreated)
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//
	//if _, ok := userCache[id]; !ok {
	//	http.NotFound(w, r)
	//}
	//
	//cacheMutex.Lock()
	//delete(userCache, id)
	//cacheMutex.Unlock()

	if err = c.userService.DeleteUser(id); err != nil {
		http.NotFound(w, r)
	}

	fmt.Println("delete user id:", id)
	w.WriteHeader(http.StatusNoContent)
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	fmt.Println("get user id:", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//cacheMutex.RLock()
	//user, ok := userCache[id]
	//cacheMutex.RUnlock()

	//if !ok {
	//	http.NotFound(w, r)
	//}

	user, err := c.userService.GetUser(id)
	fmt.Println("get user id:", user)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(j)
}
