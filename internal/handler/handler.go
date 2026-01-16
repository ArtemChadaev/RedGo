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

func (h *Handler) Routes(apiKey string) *gin.Engine {
	router := gin.New()

	api := router.Group("/api/v1")
	{
		// Группа с защитой API-ключом
		incident := api.Group("/incidents", h.apiKeyMiddleware(apiKey))
		{
			incident.POST("/", h.createIncident)
			incident.GET("/", h.getIncidents)
			incident.GET("/:id", h.getIncidentByID)
			incident.PUT("/:id", h.updateIncident)
			incident.DELETE("/:id", h.deleteIncident)
			incident.GET("/stats", h.getStats)
		}

		api.POST("/location/check", h.checkLocation)
		api.GET("/system/health", h.healthCheck)
	}

	return router
}
