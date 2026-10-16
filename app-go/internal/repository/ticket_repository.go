package repository

import (
	"benchmark-go/internal/model"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Ticket, error)
	FindAll(ctx context.Context, limit, offset int) ([]model.Ticket, error)
	Insert(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error)
	GetStatistics(ctx context.Context) (*model.TicketStatistics, error)
	FindRecentForIoBenchmark(ctx context.Context, limit int) ([]model.Ticket, error)
}

type ticketRepository struct {
	pool *pgxpool.Pool
}

func NewTicketRepository(pool *pgxpool.Pool) TicketRepository {
	return &ticketRepository{
		pool: pool,
	}
}

func (r *ticketRepository) ensureContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), 3000*time.Second)
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		return context.WithTimeout(ctx, 3000*time.Second)
	}
	return ctx, func() {}
}

func (r *ticketRepository) FindByID(ctx context.Context, id int64) (*model.Ticket, error) {
	reqCtx, cancel := r.ensureContextTimeout(ctx)
	defer cancel()

	var t model.Ticket
	query := "SELECT id, title, description, status, priority, assignee, created_at, updated_at FROM tickets WHERE id = $1"
	err := r.pool.QueryRow(reqCtx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Assignee, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *ticketRepository) FindAll(ctx context.Context, limit, offset int) ([]model.Ticket, error) {
	reqCtx, cancel := r.ensureContextTimeout(ctx)
	defer cancel()

	query := "SELECT id, title, description, status, priority, assignee, created_at, updated_at FROM tickets ORDER BY id DESC LIMIT $1 OFFSET $2"
	rows, err := r.pool.Query(reqCtx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []model.Ticket{}
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Assignee, &t.CreatedAt, &t.UpdatedAt); err == nil {
			tickets = append(tickets, t)
		}
	}
	return tickets, nil
}

func (r *ticketRepository) Insert(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error) {
	reqCtx, cancel := r.ensureContextTimeout(ctx)
	defer cancel()

	now := time.Now()
	status := "OPEN"
	var newID int64

	query := "INSERT INTO tickets (title, description, status, priority, assignee, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	err := r.pool.QueryRow(reqCtx, query, req.Title, req.Description, status, req.Priority, req.Assignee, now, now).Scan(&newID)
	if err != nil {
		return nil, err
	}

	return &model.Ticket{
		ID:          newID,
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    req.Priority,
		Assignee:    req.Assignee,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (r *ticketRepository) GetStatistics(ctx context.Context) (*model.TicketStatistics, error) {
	reqCtx, cancel := r.ensureContextTimeout(ctx)
	defer cancel()

	var s model.TicketStatistics
	query := `SELECT
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE status = 'OPEN') AS open,
		COUNT(*) FILTER (WHERE status = 'IN_PROGRESS') AS in_progress,
		COUNT(*) FILTER (WHERE status = 'RESOLVED') AS resolved,
		COUNT(*) FILTER (WHERE status = 'CLOSED') AS closed,
		COUNT(*) FILTER (WHERE priority = 'CRITICAL') AS critical,
		COUNT(*) FILTER (WHERE priority = 'HIGH') AS high,
		COUNT(*) FILTER (WHERE priority = 'MEDIUM') AS medium,
		COUNT(*) FILTER (WHERE priority = 'LOW') AS low
		FROM tickets`

	err := r.pool.QueryRow(reqCtx, query).Scan(&s.Total, &s.Open, &s.InProgress, &s.Resolved, &s.Closed, &s.Critical, &s.High, &s.Medium, &s.Low)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ticketRepository) FindRecentForIoBenchmark(ctx context.Context, limit int) ([]model.Ticket, error) {
	reqCtx, cancel := r.ensureContextTimeout(ctx)
	defer cancel()

	query := "SELECT id, title, description, status, priority, assignee, created_at, updated_at FROM tickets ORDER BY created_at DESC LIMIT $1"
	rows, err := r.pool.Query(reqCtx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []model.Ticket{}
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Assignee, &t.CreatedAt, &t.UpdatedAt); err == nil {
			tickets = append(tickets, t)
		}
	}
	return tickets, nil
}
