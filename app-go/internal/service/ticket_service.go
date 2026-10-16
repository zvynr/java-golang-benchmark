package service

import (
	"benchmark-go/internal/model"
	"benchmark-go/internal/repository"
	"context"
)

type TicketService interface {
	GetTicketByID(ctx context.Context, id int64) (*model.Ticket, error)
	GetAllTickets(ctx context.Context, page, size int) ([]model.Ticket, error)
	CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error)
	GetStatistics(ctx context.Context) (*model.TicketStatistics, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
}

func NewTicketService(ticketRepo repository.TicketRepository) TicketService {
	return &ticketService{ticketRepo: ticketRepo}
}

func (s *ticketService) GetTicketByID(ctx context.Context, id int64) (*model.Ticket, error) {
	return s.ticketRepo.FindByID(ctx, id)
}

func (s *ticketService) GetAllTickets(ctx context.Context, page, size int) ([]model.Ticket, error) {
	if size <= 0 {
		size = 20
	}
	offset := page * size
	return s.ticketRepo.FindAll(ctx, size, offset)
}

func (s *ticketService) CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error) {
	return s.ticketRepo.Insert(ctx, req)
}

func (s *ticketService) GetStatistics(ctx context.Context) (*model.TicketStatistics, error) {
	return s.ticketRepo.GetStatistics(ctx)
}
