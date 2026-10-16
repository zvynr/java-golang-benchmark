package controller

import (
	"benchmark-go/internal/model"
	"benchmark-go/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type TicketController struct {
	ticketService service.TicketService
}

func NewTicketController(ticketService service.TicketService) *TicketController {
	return &TicketController{ticketService: ticketService}
}

func (t *TicketController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id", t.GetByID)
	rg.GET("", t.GetAll)
	rg.POST("", t.Create)
	rg.GET("/statistics", t.GetStatistics)
}

func (t *TicketController) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	ticket, err := t.ticketService.GetTicketByID(c.Request.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		RespondError(c, http.StatusNotFound, "Ticket not found")
		return
	}
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(c, http.StatusOK, ticket)
}

func (t *TicketController) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	tickets, err := t.ticketService.GetAllTickets(c.Request.Context(), page, size)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(c, http.StatusOK, tickets)
}

func (t *TicketController) Create(c *gin.Context) {
	var req model.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	ticket, err := t.ticketService.CreateTicket(c.Request.Context(), &req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(c, http.StatusCreated, ticket)
}

func (t *TicketController) GetStatistics(c *gin.Context) {
	stats, err := t.ticketService.GetStatistics(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(c, http.StatusOK, stats)
}
