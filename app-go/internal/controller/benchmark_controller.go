package controller

import (
	"benchmark-go/internal/model"
	"benchmark-go/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BenchmarkController struct {
	benchmarkService service.BenchmarkService
}

func NewBenchmarkController(benchmarkService service.BenchmarkService) *BenchmarkController {
	return &BenchmarkController{benchmarkService: benchmarkService}
}

func (b *BenchmarkController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/cpu", b.Cpu)
	rg.GET("/io", b.Io)
	rg.GET("/concurrency", b.Concurrency)
	rg.POST("/allocation", b.Allocation)
}

func (b *BenchmarkController) Cpu(c *gin.Context) {
	var req model.CpuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Iterations = 10000
		req.HashRounds = 5
		req.SortSize = 5000
	}
	if req.Iterations <= 0 {
		req.Iterations = 10000
	}
	if req.HashRounds <= 0 {
		req.HashRounds = 5
	}
	if req.SortSize <= 0 {
		req.SortSize = 5000
	}

	res, err := b.benchmarkService.CpuBound(c.Request.Context(), req.Iterations, req.HashRounds, req.SortSize)
	if err != nil {
		RespondError(c, http.StatusRequestTimeout, err.Error())
		return
	}
	RespondJSON(c, http.StatusOK, res)
}

func (b *BenchmarkController) Io(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	tickets, err := b.benchmarkService.IoBound(c.Request.Context(), limit)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(c, http.StatusOK, gin.H{
		"count":   len(tickets),
		"tickets": tickets,
	})
}

func (b *BenchmarkController) Concurrency(c *gin.Context) {
	res, err := b.benchmarkService.ConcurrencyBound(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	RespondJSON(c, http.StatusOK, res)
}

func (b *BenchmarkController) Allocation(c *gin.Context) {
	var req model.AllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.ObjectCount = 10000
		req.ObjectSizeBytes = 256
		req.TransformRounds = 3
	}

	res, err := b.benchmarkService.AllocationBound(c.Request.Context(), req.ObjectCount, req.ObjectSizeBytes, req.TransformRounds)
	if err != nil {
		RespondError(c, http.StatusRequestTimeout, err.Error())
		return
	}
	RespondJSON(c, http.StatusOK, res)
}
