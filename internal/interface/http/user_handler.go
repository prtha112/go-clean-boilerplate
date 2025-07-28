package http

import (
    "encoding/json"
    "net/http"
    "strconv"

    domain "go-clean-architecture/internal/domain/user"
    "github.com/gorilla/mux"
)

type UserHandler struct {
    useCase domain.UseCase
}

func NewUserHandler(r *mux.Router, uc domain.UseCase) {
    handler := &UserHandler{useCase: uc}
    r.HandleFunc("/users/{id}", handler.GetUser).Methods("GET")
    r.HandleFunc("/users", handler.CreateUser).Methods("POST")
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    user, err := h.useCase.GetUser(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user domain.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    if err := h.useCase.CreateUser(&user); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
