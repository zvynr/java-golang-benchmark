# Java 21 JVM vs GraalVM Native 21 vs Go Performance Benchmark Suite

![Java](https://img.shields.io/badge/Java-21-orange?style=for-the-badge&logo=openjdk)
![GraalVM](https://img.shields.io/badge/GraalVM-Native%20Image-blue?style=for-the-badge&logo=graalvm)
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go)
![Spring Boot](https://img.shields.io/badge/Spring%20Boot-3.3%2B-6DB33F?style=for-the-badge&logo=springboot)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)
![Grafana k6](https://img.shields.io/badge/Grafana-k6-7D64FF?style=for-the-badge&logo=k6)
![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-E6522C?style=for-the-badge&logo=prometheus)

A high-performance benchmarking architecture designed to rigorously evaluate and compare **Java 21 HotSpot JVM (with Virtual Threads)**, **GraalVM Native Image 21 (Ahead-Of-Time AOT)**, and **Go (Golang)** across throughput, cold-start latency, memory footprint (RSS/Heap), and CPU efficiency under simulated high-concurrency workloads.

---

## 🛠️ Tech Stack & Core Technologies

| Category | Technologies & Libraries |
| :--- | :--- |
| **Languages & Runtimes** | Java 21 (OpenJDK HotSpot), GraalVM Native Image 21, Go (Golang) 1.22+ |
| **Frameworks** | Spring Boot 3.3+ (with Virtual Threads enabled), Spring Web, Spring Data JPA, Gin (Go Web Framework) |
| **Database & ORM** | PostgreSQL 16.15, HikariCP Connection Pool, `pgx`/`pgxpool`, Flyway Database Migrations |
| **Load Testing** | Grafana k6 (Custom load testing scripts `w0` - `w7`) |
| **Observability** | Prometheus, Promtail, cAdvisor, Grafana 11 (Pre-provisioned dashboards) |
| **Orchestration** | Docker, Docker Compose (Multi-service profiles) |

---

## 🏗️ Architectural Overview

### Visual System Flow (ASCII Architecture Diagram)
```text
  [ Grafana k6 Load Generator Suite ]
                  │
                  ▼  HTTP / REST API Requests
  ┌─────────────────────────────────────────────────────────────┐
  │                 Docker Compose Services                     │
  │  ├─ Spring Boot Java 21 JVM (+ Virtual Threads) (Port 8080) │
  │  ├─ GraalVM Native Image 21 (+ Virtual Threads) (Port 8081) │
  │  └─ Go Gin Service                              (Port 8082) │
  └──────────────────────────────┬──────────────────────────────┘
                                 │
                                 ▼  JDBC / pgxpool Database Transactions
                    [ PostgreSQL 16.15 Database ]
                                 │
                                 ▼  Metrics & Telemetry Scraping
         [ Prometheus, cAdvisor, Promtail & Grafana Dashboards ]
```

### Detailed Mermaid Architecture Diagram
```mermaid
graph TD
    subgraph Client Load Generator
        K6[Grafana k6 Load Test Suite]
    end

    subgraph API Gateways / Services (Docker Profiles)
        JVM21[Spring Boot Java 21 JVM<br/>Port 8080]
        Native21[GraalVM Native 21<br/>Port 8081]
        GoSvc[Go Gin Service<br/>Port 8082]
    end

    subgraph Data & Persistence Layer
        PG[(PostgreSQL 16.15<br/>HikariCP / pgxpool)]
    end

    subgraph Observability Stack
        PROM[Prometheus Metrics]
        CADV[cAdvisor Metrics]
        GRAF[Grafana Dashboards]
    end

    K6 -->|HTTP / REST| JVM21
    K6 -->|HTTP / REST| Native21
    K6 -->|HTTP / REST| GoSvc

    JVM21 --> PG
    Native21 --> PG
    GoSvc --> PG

    JVM21 -.-> PROM
    Native21 -.-> PROM
    GoSvc -.-> PROM
    CADV -.-> PROM
    PROM --> GRAF
```

---

## 📋 Testing Plan & Workload Scenarios

| Workload | Keterangan Workload | Proses API Test & Skenario Pengujian |
| :--- | :--- | :--- |
| **W0** | **Cold Start & Size** | Mengukur seberapa cepat aplikasi siap melayani request pertama (*Time-to-First-Request*) serta ukuran kontainer runtime. |
| **W1** | **CPU Bound** | Mengukur performa saat memproses *CPU-bound task* berat:<br>• 10.000x SHA-256 Hashing<br>• Array Sorting (5.000 int)<br>• Agregasi Median & Sum<br>• **Skenario**: VU = 1, 10, 50, 100 (Jeda 15 detik antar VU). |
| **W2** | **I/O Bound** | Mengukur perbedaan latensi saat aplikasi banyak menunggu *external I/O* (Database query):<br>• Mengambil list 20 data terbaru dari PostgreSQL<br>• **Skenario**: VU = 10, 50, 100 (Jeda 15 detik antar VU). |
| **W3** | **High Concurrency** | Menguji batas ketahanan saat beban *Virtual Users* (VU) dinaikkan bertahap:<br>• CPU-bound ringan (looping 500) + `findById()`<br>• **Skenario**: VU = 10, 50, 100, 200 (Jeda 15 detik antar VU). |
| **W4** | **Allocation & GC** | Menguji tekanan alokasi objek sementara dan performa *Garbage Collection* (GC):<br>• Alokasi 10.000 *byte array* + 3x *transform round* (penciptaan ribuan objek sementara)<br>• **Skenario**: VU = 10, 50, 100 (Jeda 15 detik antar VU). |
| **W5** | **Realistic Test (CRUD Mix)** | Simulasi aplikasi bisnis nyata dengan variasi transaksi:<br>• ~40% `getById`<br>• ~30% `getListPagination`<br>• ~20% `create` + validasi<br>• ~10% query agregasi statistik<br>• **Skenario**: VU 100, 300 (60s warm-up, 5m stabil, 30s ramp-down). |
| **W6** | **Realistic Test 2 (Heavy Mix)** | Simulasi aplikasi bisnis nyata dengan variasi transaksi serta proses berat di CPU & alokasi:<br>• ~5% `cpuBound`, ~5% `allocationBound`, ~30% `getById`, ~30% `getListPagination`, ~20% `create` + validasi, ~10% statistik<br>• **Skenario**: VU 100 (60s warm-up, 5m stabil, 30s ramp-down). |

---

## ⚙️ System & Runtime Configurations

### 1. Spesifikasi Sistem & Resource Constraints
| No | Konfigurasi | Nilai | Keterangan |
| :---: | :--- | :--- | :--- |
| **1** | Container CPU | 2 Cores | Batas *hard limit* CPU kernel Linux (CFS Throttling) seragam untuk seluruh runtime. |
| **2** | Container RAM | 2 GB | Batas kapasitas memori fisik maksimum seragam untuk seluruh runtime. |
| **3** | Database Engine | PostgreSQL 16.15 Alpine | Relational database backend. |
| **4** | Database Connection Pool | Max = 20, Min Idle = 5, Conn Timeout = 30s, Idle Timeout = 10m, Max Lifetime = 30m | Konfigurasi kolam koneksi (HikariCP / `pgxpool`) yang diseragamkan antar framework. |

### 2. Spesifikasi Runtime & Framework Aplikasi
| Bahasa & Mode Executable | Runtime / Version | Threading & Memory Model | Pustaka & Framework Utama |
| :--- | :--- | :--- | :--- |
| **Java 21 - JVM HotSpot** | Eclipse Temurin JRE 21 | G1GC / ZGC<br>**Virtual Threads Active** | • Framework: Spring Boot 3.3.5 (`spring-boot-starter-web`)<br>• DB Pool: HikariCP (`spring-boot-starter-jdbc`)<br>• DB Driver: PostgreSQL JDBC (`org.postgresql:postgresql`)<br>• Metrics: Micrometer Prometheus Registry (`micrometer-registry-prometheus`)<br>• Validation: Spring Boot Validation (`spring-boot-starter-validation`)<br>• Utils: Lombok |
| **Java 21 - Native Image** | GraalVM CE 21 (AOT) | Substrate VM Thread Pool + Serial GC<br>**Virtual Threads Active** | • Native Plugin: `native-maven-plugin:0.10.3`<br>• Optimasi Build: `-O2 --no-fallback -H:+ReportExceptionStackTraces` |
| **Golang - Binary Native** | Go 1.22 (Alpine) | Concurrent Goroutines + Go Concurrent GC | • HTTP Router: Gin v1.10.0 (`github.com/gin-gonic/gin`)<br>• DB Pool/Driver: `pgxpool` (`github.com/jackc/pgx/v5`)<br>• Metrics: Prometheus Go Client (`github.com/prometheus/client_golang`) |

### 3. Stack Monitoring & Observability
| Tool | Versi Kontainer | Peran & Deskripsi |
| :--- | :--- | :--- |
| **k6** | `grafana/k6:0.54.0` | **Load Testing Engine**: Menyuntikkan beban *Virtual Users* (VU) otomatis berbasis skrip JavaScript (`W0`–`W6`). |
| **Prometheus** | `prom/prometheus:v2.54.1` | **Time-Series Database (TSDB)**: Menarik (*scrape interval 5s*) & menyimpan metrik dari Actuator, cAdvisor, & Promtail. |
| **cAdvisor** | `google/cadvisor:v0.49.1` | **Container Resource Monitoring**: Merekam statistik kernel Linux CPU Cores, CFS Throttling Rate, dan RAM Working Set (RSS). |
| **Promtail** | `grafana/promtail:2.9.2` | **Log Parser & Shipper**: Memindai log `-XX:+PrintGC` dari container dan mengubahnya menjadi metrik angka kustom Prometheus. |
| **Grafana** | `grafana/grafana-oss:11.2.0` | **Master Visualization Engine**: Dashboard interaktif terpusat untuk menampilkan grafik RPS, Latency p95, Error Rate %, RAM, CPU, dan GC *side-by-side*. |

---

## 🔒 Security & Environment Management

This repository adheres to strict **production-ready security best practices**:
1. **Zero Secret Exposure**: No sensitive credentials (database passwords, API tokens, Grafana admin passwords) are hardcoded in version-controlled configuration files (`docker-compose.yml`, `application.yml`).
2. **Environment Template (`.env.example`)**: A fully documented `.env.example` file is provided. Developers and CI/CD pipelines must copy `.env.example` to `.env` and configure environment-specific values.
3. **Automated `.gitignore` Protection**: Local `.env` and `.env.local` files are strictly ignored by Git to prevent accidental credential leaks.

### Quick Environment Setup
```bash
# 1. Copy the environment template
cp .env.example .env

# 2. (Optional) Edit .env with your custom credentials
nano .env
```

---

## 🚀 Getting Started

### Prerequisites
- **Docker Engine** (24.0+) & **Docker Compose** (v2+)
- **Java 21 JDK** & Maven 3.9+ (if building Java apps locally)
- **Go 1.22+** (if building Go app locally)
- **k6** (Grafana k6 load testing tool)

### 1. Initialize Docker Network and Volumes
```bash
docker network create benchmark-net
docker volume create pgdata
```

### 2. Run the Benchmark (`benchmark.sh`)
Use the standalone shell orchestration script (`benchmark.sh`) to automatically run the entire benchmark pipeline across runtimes and workloads:
```bash
# Make script executable
chmod +x benchmark.sh

# Run all workloads across all runtimes (Java 21 JVM, Native 21, and Go)
./benchmark.sh all

# Or run specific runtime and workload
RUNTIME=jvm ./benchmark.sh W1
RUNTIME=native REPEATS=3 ./benchmark.sh W3
RUNTIME=go ./benchmark.sh W5
```

Alternatively, run specific services using Docker Compose profiles:
```bash
# Start PostgreSQL and Prometheus/Grafana monitoring stack
docker compose --profile monitoring up -d postgres prometheus grafana cadvisor promtail

# Run Java 21 JVM Service
docker compose --profile jvm up -d app-jvm

# Run GraalVM Native 21 Service
docker compose --profile native up -d app-native

# Run Go High-Performance Service
docker compose --profile go up -d app-go
```

---

## 📈 Monitoring & Observability

Access the pre-configured observability dashboards via Grafana:
- **URL**: [http://localhost:3000](http://localhost:3000)
- **Default Username**: `admin` (configurable via `GF_SECURITY_ADMIN_USER` in `.env`)
- **Default Password**: `admin` (configurable via `GF_SECURITY_ADMIN_PASSWORD` in `.env`)

Dashboards included:
1. **Overview (`jvm-vs-native.json`)**: High-level throughput, latency percentiles (P50, P95, P99), and error rates across runtimes.
2. **Resource Utilization (`jvm-vs-native-overview.json`)**: CPU usage, Resident Set Size (RSS) memory consumption, and GC pause metrics.
3. **Deep Dive Metrics (`jvm-vs-native-details.json`)**: HikariCP / pgxpool connection pool utilization, thread states, and system-level cAdvisor telemetry.

---

## 📋 Project Directory Structure

```tree
java-jvm-vs-native-benchmark/
├── .env.example                # Environment variable template
├── .env                        # Local environment variables (Git-ignored)
├── .gitignore                  # Git exclusion rules (ignoring .env, target/, results/)
├── docker-compose.yml          # Multi-service orchestration config
├── benchmark.sh                # Standalone shell benchmark orchestrator
├── app/                        # Spring Boot (Java 21) Application
│   ├── Dockerfile.jvm21        # OpenJDK 21 JVM + Virtual Threads container definition
│   ├── Dockerfile.native21     # GraalVM Native Image 21 container definition
│   ├── pom.xml                 # Maven build configuration & GraalVM native plugins
│   └── src/main/               # Spring Boot controllers, services, repositories
├── app-go/                     # Go (Golang) High-Performance Service
│   ├── Dockerfile.go           # Multi-stage optimized Go scratch/alpine container
│   ├── go.mod                  # Go modules (Gin, pgxpool)
│   ├── cmd/server/main.go      # Entrypoint & graceful shutdown
│   └── internal/               # Clean architecture (controllers, services, repositories, worker)
├── k6/                         # k6 Load Testing Scripts (w0 to w7)
└── monitoring/                 # Prometheus, Promtail, & Grafana Provisioning Dashboards
```

---

## 📜 License

This project is licensed under the **MIT License** - see the LICENSE file for details.
