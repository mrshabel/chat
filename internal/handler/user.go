package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/repository"
	"github.com/mrshabel/chat/internal/service"
	"github.com/mrshabel/chat/internal/util"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Create(r.Context(), &req)
	if err != nil {
		log.Println(err)
		if errors.Is(err, service.ErrUserAlreadyExist) {
			util.WriteError(w, "Username already taken", http.StatusConflict)
			return
		}
		util.WriteError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, user, http.StatusCreated)
}

func (h *UserHandler) GetByUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := util.GetParamUUID(r, "id")
	if err != nil {
		util.WriteError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			util.WriteError(w, "User not found", http.StatusNotFound)
			return
		}
		util.WriteError(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, user, http.StatusOK)
}
