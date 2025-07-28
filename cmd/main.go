package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	orderRepo "go-clean-architecture/internal/infrastructure/order"
	userRepo "go-clean-architecture/internal/infrastructure/user"
	httpHandler "go-clean-architecture/internal/interface/http"
	orderUsecase "go-clean-architecture/internal/usecase/order"
	userUsecase "go-clean-architecture/internal/usecase/user"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=15432 user=mock password=mock123 dbname=mockdb sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	router := httpHandler.NewRouter()

	userRepo := userRepo.NewUserRepository()
	userUC := userUsecase.NewUserUseCase(userRepo)
	httpHandler.NewUserHandler(router, userUC)

	orderRepository := orderRepo.NewPostgresOrderRepository(db)
	orderUC := orderUsecase.NewOrderUseCase(orderRepository)
	httpHandler.NewOrderHandler(router, orderUC)

	log.Println("Listening on :8085")
	log.Fatal(http.ListenAndServe(":8085", router))
}
