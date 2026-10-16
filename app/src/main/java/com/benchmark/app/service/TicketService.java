package com.benchmark.app.service;

import com.benchmark.app.model.Ticket;
import com.benchmark.app.request.CreateTicketRequest;
import com.benchmark.app.response.TicketStatistics;
import com.benchmark.app.repository.TicketRepository;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class TicketService {

    private final TicketRepository ticketRepository;

    public TicketService(TicketRepository ticketRepository) {
        this.ticketRepository = ticketRepository;
    }

    public Optional<Ticket> getById(Long id) {
        return ticketRepository.findById(id);
    }

    public List<Ticket> list(int page, int size) {
        int limit = Math.min(Math.max(size, 1), 100);
        int offset = Math.max(page, 0) * limit;
        return ticketRepository.findAll(limit, offset);
    }

    public Ticket create(CreateTicketRequest request) {
        Ticket ticket = new Ticket();
        ticket.setTitle(request.title());
        ticket.setDescription(request.description());
        ticket.setPriority(request.priority() != null ? request.priority().toUpperCase() : "MEDIUM");
        ticket.setAssignee(request.assignee());
        ticket.setStatus("OPEN");
        return ticketRepository.insert(ticket);
    }

    public TicketStatistics statistics() {
        return ticketRepository.getStatistics();
    }
}
