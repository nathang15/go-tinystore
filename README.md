# Distributed cache implementation 

## ⚠️This is for learning purposes so the implementation might not be correct, robust, or production applicable.

### Features:
- LRU cache
- Consistent hashing implementation uses the concept of virtual nodes. Users can specify the virtual nodes size when initializing the consistent hash ring. 
- Note that this is a very unfair distribution for virtual nodes size lesser than 100. The distribution becomes gradually consistent when virtual nodes size are increased, it seems most consistent if the amount of vnodes is greater than 700. See the output.txt file.
- Bully algorithm for leader election of cluster.
- Dynamic node can join/leave cluster and every other config in consistent hashing and leader will be updated instanteneously.
- Therefore, it has no single point of failure as there is always guaranteed to have a leader.
- 
#### How to run:
- docker-compose build -> docker-compose up -> docker build -t client -f Dockerfile.client . -> docker run --network tinystore_default client

#### Note to self:
- protoc service.proto --go_out=.     
- protoc --go-grpc_out=. service.proto

## Reference

- [Setup mTLS](https://dev.to/techschoolguru/a-complete-overview-of-ssl-tls-and-its-cryptographic-system-36pd)
- [Design](https://www.youtube.com/watch?v=iuqZvajTOyA&t=920s)
- [Consistent hash](https://www.youtube.com/watch?v=UF9Iqmg94tk&t=359s)
- [Bully algo](https://lass.cs.umass.edu/~shenoy/courses/spring22/lectures/Lec14_notes.pdf)
