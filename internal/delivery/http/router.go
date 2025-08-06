package http

import (
	"go-clean-boilerplate/internal/domain"
	"go-clean-boilerplate/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	productHandler *ProductHandler
	orderHandler   *OrderHandler
	authHandler    *AuthHandler
	invoiceHandler *InvoiceHandler
	authUsecase    domain.AuthUsecase
}

func NewRouter(productUsecase domain.ProductUsecase, orderUsecase domain.OrderUsecase, authUsecase domain.AuthUsecase, invoiceUsecase domain.InvoiceUsecase) *Router {
	return &Router{
		productHandler: NewProductHandler(productUsecase),
		orderHandler:   NewOrderHandler(orderUsecase),
		authHandler:    NewAuthHandler(authUsecase),
		invoiceHandler: NewInvoiceHandler(invoiceUsecase),
		authUsecase:    authUsecase,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.GET("/profile", middleware.JWTAuth(r.authUsecase), r.authHandler.GetProfile)
		}

		// Protected routes (require JWT)
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(r.authUsecase))
		{
			// Product routes
			products := protected.Group("/products")
			{
				products.POST("", r.productHandler.CreateProduct)
				products.GET("", r.productHandler.GetProducts)
				products.GET("/:id", r.productHandler.GetProduct)
				products.PUT("/:id", r.productHandler.UpdateProduct)
				products.DELETE("/:id", r.productHandler.DeleteProduct)
			}

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.POST("", r.orderHandler.CreateOrder)
				orders.GET("", r.orderHandler.GetOrders)
				orders.GET("/:id", r.orderHandler.GetOrder)
				orders.PATCH("/:id/status", r.orderHandler.UpdateOrderStatus)
				orders.DELETE("/:id", r.orderHandler.DeleteOrder)
			}

			// Invoice routes
			invoices := protected.Group("/invoices")
			{
				invoices.POST("", r.invoiceHandler.CreateInvoice)
				invoices.POST("/from-order/:order_id", r.invoiceHandler.CreateInvoiceFromOrder)
				invoices.GET("", r.invoiceHandler.GetInvoices)
				invoices.GET("/overdue", r.invoiceHandler.GetOverdueInvoices)
				invoices.GET("/number/:invoice_number", r.invoiceHandler.GetInvoiceByNumber)
				invoices.GET("/:id", r.invoiceHandler.GetInvoice)
				invoices.PATCH("/:id/status", r.invoiceHandler.UpdateInvoiceStatus)
				invoices.DELETE("/:id", r.invoiceHandler.DeleteInvoice)
			}
		}
	}

	return router
}
