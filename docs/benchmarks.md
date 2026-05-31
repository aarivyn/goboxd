# Benchmarks

Hardware: WSL2 Ubuntu, 12-core machine
Test: POST /run with Python "Hello World"
All runs: 100% success rate, zero failures

| Clients | req/s | p50 | p95 | p99 |
|---------|-------|-----|-----|-----|
| 1 | 31.7 | 31ms | 40ms | 40ms |
| 10 | 210.6 | 45ms | 73ms | 93ms |
| 50 | 253.9 | 192ms | 213ms | 225ms |
| 100 | 251.8 | 389ms | 423ms | 441ms |

The server queues requests under load rather than failing them.
All 1000 requests at 100 concurrent clients completed successfully. 

