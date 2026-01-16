package handler

import (
	"net/http"
	"strconv"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/gin-gonic/gin"
)

// Middleware для проверки API-ключа
func (h *Handler) apiKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerKey := c.GetHeader("X-API-KEY")
		if headerKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}

// POST /api/v1/incidents/
func (h *Handler) createIncident(c *gin.Context) {
	var input domain.Incident
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.services.IncidentService.CreateIncident(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input) // Возвращаем объект с созданным ID
}

// GET /api/v1/incidents/
func (h *Handler) getIncidents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0")) // 0 = выдать "много" по твоей просьбе

	incidents, err := h.services.IncidentService.GetIncidents(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incidents)
}

// GET /api/v1/incidents/:id
func (h *Handler) getIncidentByID(c *gin.Context) {
	id, err := strconv.Atoi("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	inc, err := h.services.IncidentService.GetIncidentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, inc)
}

// PUT /api/v1/incidents/:id
func (h *Handler) updateIncident(c *gin.Context) {
	id, err := strconv.Atoi("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input domain.Incident
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.services.IncidentService.UpdateIncident(c.Request.Context(), id, &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /api/v1/incidents/:id
func (h *Handler) deleteIncident(c *gin.Context) {
	id, err := strconv.Atoi("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.services.IncidentService.DeleteIncident(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/incidents/stats
func (h *Handler) getStats(c *gin.Context) {
	count, err := h.services.IncidentService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_count": count})
}

// POST /api/v1/location/check
func (h *Handler) checkLocation(c *gin.Context) {
	var input struct {
		UserID int     `json:"user_id" binding:"required"`
		X      float64 `json:"x" binding:"required"`
		Y      float64 `json:"y" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nearby, err := h.services.IncidentService.CheckLocation(c.Request.Context(), input.UserID, input.X, input.Y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nearby)
}

// GET /api/v1/system/health
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
