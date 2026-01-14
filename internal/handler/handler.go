package handler

import (
	"github.com/ArtemChadaev/RedGo/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	services *service.Service
	redis    *redis.Client
}

func NewHandler(services *service.Service, redis *redis.Client) *Handler {
	return &Handler{
		services: services,
		redis:    redis,
	}
}

func (h *Handler) Routes() *gin.Engine {
	router := gin.New()

	//api := router.Group("/api/v1")
	//{
	//	incident := router.Group("/incidents")
	//	{
	//		incident.POST("/")
	//		incident.GET("/")
	//		incident.GET("/:id")
	//		incident.PUT("/:id")
	//		incident.DELETE("/:id")
	//
	//		incident.GET("/stats")
	//	}
	//
	//	api.POST("/location/check")
	//
	//	api.GET("/system/health")
	//}

	return router
}
