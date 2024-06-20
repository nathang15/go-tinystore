# go tinystore

## This is for learning purposes so the implementation might not be correct or performant.

### Features:
- LRU cache
- Consistent hashing implementation uses the concept of virtual nodes. Users can specify the virtual nodes size when initializing the consistent hash ring. 
- Note that this is a very unfair distribution for virtual nodes size lesser than 100. The distribution becomes gradually consistent when virtual nodes size are increased, it seems most consistent if the amount of vnodes is greater than 700. See the output.txt file.
- Bully algorithm for leader election of cluster.
- Dynamic node can join/leave cluster and every other config in consistent hashing and leader will be updated instanteneously.
- Therefore, it has no single point of failure as there is always guaranteed to have a leader.
Distributed cache implementation:

# CREDIT

#### command to change proto:

protoc service.proto --go_out=.     
protoc --go-grpc_out=. service.proto