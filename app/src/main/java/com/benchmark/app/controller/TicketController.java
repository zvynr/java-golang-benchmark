package com.benchmark.app.controller;

import com.benchmark.app.model.Ticket;
import com.benchmark.app.request.CreateTicketRequest;
import com.benchmark.app.response.TicketStatistics;
import com.benchmark.app.service.TicketService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/tickets")
public class TicketController {

    private final TicketService ticketService;

    public TicketController(TicketService ticketService) {
        this.ticketService = ticketService;
    }

    @GetMapping("/{id}")
    public ResponseEntity<Ticket> getById(@PathVariable Long id) {
        return ticketService.getById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping
    public ResponseEntity<List<Ticket>> list(
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size) {
        return ResponseEntity.ok(ticketService.list(page, size));
    }

    @PostMapping
    public ResponseEntity<Ticket> create(@Valid @RequestBody CreateTicketRequest request) {
        Ticket created = ticketService.create(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @GetMapping("/statistics")
    public ResponseEntity<TicketStatistics> statistics() {
        return ResponseEntity.ok(ticketService.statistics());
    }
}
