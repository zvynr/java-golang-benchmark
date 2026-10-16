package com.benchmark.app.response;

public record TicketStatistics(
        long total,
        long open,
        long inProgress,
        long resolved,
        long closed,
        long critical,
        long high,
        long medium,
        long low
) {}
