# IP Address Counter (Large-Scale File Processing in Go)

High-performance Go utility for counting unique IP addresses in extremely large text files.

This project was developed as a solution to a large-scale data processing challenge involving:

- **100 GB input file**
- ~**8 billion records**
- Input size exceeding available RAM
- Strict resource constraints

The goal was to design a memory-efficient and scalable approach without relying on external big data frameworks.

---

## Problem Statement

Counting unique IP addresses is trivial when the dataset fits into memory.

It becomes non-trivial when:

- The input file size exceeds available RAM
- The number of records reaches billions
- Memory usage must remain predictable
- Processing must remain deterministic and scalable

This implementation focuses on controlled resource consumption and disk-based processing strategies.

---

## Engineering Constraints

- 100GB input file
- ~8 billion lines
- Limited RAM
- Controlled CPU usage
- High disk I/O pressure
- Deterministic output

---

## Solution Overview

The solution is based on:

### 1. Chunk-Based Processing

The input file is processed in configurable chunks:

- Each chunk size is user-defined
- Memory footprint is predictable
- Suitable for constrained environments

### 2. Worker Pool Concurrency

Parallel processing is achieved via configurable worker pool:

- Tunable CPU usage
- Controlled memory allocation
- Balance between throughput and stability

### 3. Aggregation Phase

Partial results are merged to produce final unique count.

The architecture prioritizes:

- Stability
- Predictable memory usage
- Scalability
- Disk efficiency

---

## Usage

```bash
ip_calc -f <input_file> -size <chunk_size_bytes> -w <worker_count>
```

## Parameters

| Flag    | Description         | Default            |
| ------- | ------------------- | ------------------ |
| `-f`    | Input file          | `ip_addresses`     |
| `-size` | Chunk size in bytes | `1073741824` (1GB) |
| `-w`    | Worker pool size    | `runtime.NumCPU()` |

## Example (large dataset)
```bash
.\ip_calc -size 5368709120 -w 4
```

## Test File Generator
```bash
make generator
```
```bash
./generator -a 1000000
```

## Makefile Targets

| Target        | Description                         |
| ------------- | ----------------------------------- |
| `all`         | Build project (default)             |
| `build`       | Build for current OS                |
| `build-linux` | Build for Linux/amd64               |
| `run`         | Build and run                       |
| `generator`   | Build and run test generator        |
| `lint`        | Run linter (requires golangci-lint) |
| `test`        | Run tests                           |
| `clean`       | Remove build artifacts              |
| `help`        | Show help                           |

## Complexity Considerations
**O(chunk_size × worker_count)**

## Design Philosophy

This project demonstrates:

Systems-level engineering

Working under memory constraints

Controlled concurrency in Go

Disk-oriented data processing

Balancing CPU, RAM, and I/O trade-offs

It reflects practical backend engineering under real-world constraints.

Possible Improvements

Parallel merge optimization

Streaming-based deduplication

Bloom-filter pre-filtering

Distributed extension

Benchmark profiling and optimization
---

**Benchmark**

Test environment:

- CPU: AMD Ryzen 7 7700
- Storage: NVMe M.2 SSD
- RAM: 32 GB
- OS: (add your OS if desired)

Test parameters:
```bash
ip_calc -size 5368709120 -w 4
```
**Dataset characteristics:**

~100 GB input file

~8 billion IP address records

## Result

Total processing time: ~14 minutes

Observed characteristics:

CPU utilization balanced across workers

Memory usage bounded by chunk size × worker count

Disk I/O was the primary bottleneck

No full dataset loading into memory

This demonstrates efficient disk-based processing and controlled parallelism under large-scale constraints.
```markdown
Estimated throughput: ~120 MB/s effective processing rate (including parsing and deduplication).
```
---

## Lessons Learned

Working with large-scale disk-bound workloads highlights several important engineering realities:

- **Disk I/O dominates CPU at scale.**  
  Even with multi-core CPUs, storage throughput becomes the primary bottleneck.

- **Chunk size tuning significantly impacts performance.**  
  Larger chunks reduce merge overhead but increase memory pressure.

- **Over-parallelization degrades throughput.**  
  Increasing worker count beyond optimal limits leads to I/O contention and diminishing returns.

- **Predictable resource usage is more important than raw speed.**  
  Controlled memory and CPU utilization ensure stability under heavy load.

- **Simple architectures often outperform overly complex ones.**  
  For constrained environments, deterministic and maintainable designs are preferable to framework-heavy solutions.

This project reinforced the importance of balancing CPU, memory, and storage characteristics rather than optimizing any single dimension in isolation.
