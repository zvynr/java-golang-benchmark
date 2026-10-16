package com.benchmark.app.repository;

import com.benchmark.app.model.Ticket;
import com.benchmark.app.response.TicketStatistics;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.jdbc.support.GeneratedKeyHolder;
import org.springframework.jdbc.support.KeyHolder;
import org.springframework.stereotype.Repository;

import java.sql.PreparedStatement;
import java.sql.Statement;
import java.sql.Timestamp;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Repository
public class TicketRepository {

    private final JdbcTemplate jdbcTemplate;

    private static final RowMapper<Ticket> TICKET_MAPPER = (rs, rowNum) -> {
        Ticket t = new Ticket();
        t.setId(rs.getLong("id"));
        t.setTitle(rs.getString("title"));
        t.setDescription(rs.getString("description"));
        t.setStatus(rs.getString("status"));
        t.setPriority(rs.getString("priority"));
        t.setAssignee(rs.getString("assignee"));
        Timestamp created = rs.getTimestamp("created_at");
        Timestamp updated = rs.getTimestamp("updated_at");
        t.setCreatedAt(created != null ? created.toInstant() : null);
        t.setUpdatedAt(updated != null ? updated.toInstant() : null);
        return t;
    };

    public TicketRepository(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public Optional<Ticket> findById(Long id) {
        List<Ticket> list = jdbcTemplate.query(
                "SELECT id, title, description, status, priority, assignee, created_at, updated_at FROM tickets WHERE id = ?",
                TICKET_MAPPER, id);
        return list.isEmpty() ? Optional.empty() : Optional.of(list.get(0));
    }

    public List<Ticket> findAll(int limit, int offset) {
        return jdbcTemplate.query(
                "SELECT id, title, description, status, priority, assignee, created_at, updated_at " +
                "FROM tickets ORDER BY id DESC LIMIT ? OFFSET ?",
                TICKET_MAPPER, limit, offset);
    }

    public Ticket insert(Ticket ticket) {
        KeyHolder keyHolder = new GeneratedKeyHolder();
        Instant now = Instant.now();
        jdbcTemplate.update(connection -> {
            PreparedStatement ps = connection.prepareStatement(
                    "INSERT INTO tickets (title, description, status, priority, assignee, created_at, updated_at) " +
                    "VALUES (?, ?, ?, ?, ?, ?, ?)",
                    Statement.RETURN_GENERATED_KEYS);
            ps.setString(1, ticket.getTitle());
            ps.setString(2, ticket.getDescription());
            ps.setString(3, ticket.getStatus() != null ? ticket.getStatus() : "OPEN");
            ps.setString(4, ticket.getPriority());
            ps.setString(5, ticket.getAssignee());
            ps.setTimestamp(6, Timestamp.from(now));
            ps.setTimestamp(7, Timestamp.from(now));
            return ps;
        }, keyHolder);

        List<Map<String, Object>> keyList = keyHolder.getKeyList();
        if (keyList != null && !keyList.isEmpty()) {
            Map<String, Object> keys = keyList.get(0);
            Object idVal = keys.get("id");
            if (idVal instanceof Number) {
                ticket.setId(((Number) idVal).longValue());
            }
        }
        ticket.setCreatedAt(now);
        ticket.setUpdatedAt(now);
        if (ticket.getStatus() == null) {
            ticket.setStatus("OPEN");
        }
        return ticket;
    }

    public TicketStatistics getStatistics() {
        return jdbcTemplate.queryForObject(
                "SELECT " +
                "COUNT(*) AS total, " +
                "COUNT(*) FILTER (WHERE status = 'OPEN') AS open, " +
                "COUNT(*) FILTER (WHERE status = 'IN_PROGRESS') AS in_progress, " +
                "COUNT(*) FILTER (WHERE status = 'RESOLVED') AS resolved, " +
                "COUNT(*) FILTER (WHERE status = 'CLOSED') AS closed, " +
                "COUNT(*) FILTER (WHERE priority = 'CRITICAL') AS critical, " +
                "COUNT(*) FILTER (WHERE priority = 'HIGH') AS high, " +
                "COUNT(*) FILTER (WHERE priority = 'MEDIUM') AS medium, " +
                "COUNT(*) FILTER (WHERE priority = 'LOW') AS low " +
                "FROM tickets",
                (rs, rowNum) -> new TicketStatistics(
                        rs.getLong("total"),
                        rs.getLong("open"),
                        rs.getLong("in_progress"),
                        rs.getLong("resolved"),
                        rs.getLong("closed"),
                        rs.getLong("critical"),
                        rs.getLong("high"),
                        rs.getLong("medium"),
                        rs.getLong("low")
                ));
    }

    /** Simple fixed query used by W2 I/O benchmark */
    public List<Ticket> findRecentForIoBenchmark(int limit) {
        return jdbcTemplate.query(
                "SELECT id, title, description, status, priority, assignee, created_at, updated_at " +
                "FROM tickets ORDER BY created_at DESC LIMIT ?",
                TICKET_MAPPER, limit);
    }
}
