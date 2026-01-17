package handler

import (
	"github.com/ArtemChadaev/RedGo/internal/service"
	"github.com/ArtemChadaev/RedGo/internal/worker"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Service
	worker   *worker.WebhookWorker
}

func NewHandler(services *service.Service, worker *worker.WebhookWorker) *Handler {
	return &Handler{
		services: services,
		worker:   worker,
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
