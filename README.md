# Distributed cache implementation 

## ⚠️This is for learning purposes so the implementation might not be correct, robust, or production applicable.

### Features:
- LRU cache
- Consistent hashing implementation uses the concept of virtual nodes. Devs can specify the virtual nodes size when initializing the consistent hash ring. 
- Note that this is a very unfair distribution for virtual nodes size lesser than 100. The distribution becomes gradually consistent when virtual nodes size are increased, it seems most consistent if the amount of vnodes is greater than 700. See [output.txt](https://github.com/nathang15/go-tinystore/blob/main/output.txt)
- Bully algorithm for leader election of cluster
- Dynamic node can join/leave cluster and every other config in consistent hashing and leader will be updated instanteneously. Therefore, it has no single point of failure as there is always guaranteed to have a leader.

### Performance:
#### With Docker containerized cache servers
```
=== RUN   Test10kPutsNoVnode
    client_test.go:52: Time to complete 10k puts with REST API, no virtual node: 6.145886283s
    client_test.go:53: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kPutsNoVnode (6.20s)
=== RUN   Test10kPuts
    client_test.go:88: Time to complete 10k puts with REST API, 10 virtual nodes: 4.041716693s
    client_test.go:89: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kPuts (4.14s)
=== RUN   Test50kPuts
    client_test.go:124: Time to complete 50k puts with REST API, 10 virtual nodes: 3.685566418s
    client_test.go:125: Cache misses: 0/50,000 (0.000000%)
--- PASS: Test50kPuts (3.74s)
=== RUN   Test10kGRPCPutsNoVnode
    client_test.go:188: Time to complete 10k puts with GRPC, no virtual node: 1.21835518s
    client_test.go:189: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kGRPCPutsNoVnode (1.29s)
=== RUN   Test10kGRPCPuts
    client_test.go:223: Time to complete 10k puts with GRPC, 10 virtual nodes: 1.436099935s
    client_test.go:224: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kGRPCPuts (1.50s)
=== RUN   Test50kGRPCPuts
    client_test.go:259: Time to complete 50k puts with GRPC, 10 virtual nodes: 3.32016691s
    client_test.go:260: Cache misses: 0/50,000 (0.000000%)
--- PASS: Test50kGRPCPuts (3.36s)
```
#### On local machine
```
=== RUN   Test10kRestApiPuts
    client_test.go:52: Time to complete 10k puts with REST API, 0 virtual node: 4.9028426s
    client_test.go:53: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kPutsNoVnode (6.20s)
=== RUN   Test10kRestApiPuts10Vnode
    client_test.go:88: Time to complete 10k puts with REST API, 10 virtual nodes: 3.0234218s
    client_test.go:89: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kPuts (4.14s)
=== RUN   Test50kPuts
    client_test.go:124: Time to complete 50k puts with REST API, 10 virtual nodes: 3.7663467s
    client_test.go:125: Cache misses: 0/50,000 (0.000000%)
--- PASS: Test50kPuts (3.74s)
=== RUN   Test10kGrpcPuts
    client_test.go:188: Time to complete 10k puts with gRPC, no virtual node: 590.634661ms
    client_test.go:189: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kGRPCPutsNoVnode (1.29s)
=== RUN   Test10kGRPCPuts
    client_test.go:223: Time to complete 10k puts with gRPC, 10 virtual nodes: 472.409995ms
    client_test.go:224: Cache misses: 0/10,000 (0.000000%)
--- PASS: Test10kGRPCPuts (1.50s)
=== RUN   Test50kGRPCPuts
    client_test.go:259: Time to complete 50k puts with gRPC, 10 virtual nodes: 529.603942ms
    client_test.go:260: Cache misses: 0/50,000 (0.000000%)
--- PASS: Test50kGRPCPuts (3.36s)
```
## Reference

- [Setup mTLS](https://dev.to/techschoolguru/a-complete-overview-of-ssl-tls-and-its-cryptographic-system-36pd)
- [Design](https://www.youtube.com/watch?v=iuqZvajTOyA&t=920s)
- [Consistent hash](https://www.youtube.com/watch?v=UF9Iqmg94tk&t=359s)
- [Bully algo](https://lass.cs.umass.edu/~shenoy/courses/spring22/lectures/Lec14_notes.pdf)

#### Note to self:
- `protoc service.proto --go_out=.`     
- `protoc --go-grpc_out=. service.proto`
- `docker-compose build` -> `docker-compose up` -> `docker image build -t client -f Dockerfile.client .` -> `docker run --network tinystore_default client`
