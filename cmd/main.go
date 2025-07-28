package main

import (
	"log"
	"net/http"

	userRepo "go-clean-architecture/internal/infrastructure/user"
	httpHandler "go-clean-architecture/internal/interface/http"
	userUsecase "go-clean-architecture/internal/usecase/user"
)

func main() {
	router := httpHandler.NewRouter()

	userRepo := userRepo.NewUserRepository()
	userUC := userUsecase.NewUserUseCase(userRepo)
	httpHandler.NewUserHandler(router, userUC)

	log.Println("Listening on :8085")
	log.Fatal(http.ListenAndServe(":8085", router))
}
