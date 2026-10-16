package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) RegisterRoutes(r *gin.Engine) {
	r.GET("/actuator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	promHandler := gin.WrapH(promhttp.Handler())
	r.GET("/actuator/prometheus", promHandler)
	r.GET("/metrics", promHandler)
}
