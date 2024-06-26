package ch

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nathang15/go-tinystore/internal/node"
)

type Ring struct {
	Nodes      node.Nodes
	Virtual    int
	VirtualMap map[string]string
	sync.RWMutex
}

func InitRing(virtual int) *Ring {
	return &Ring{Nodes: node.Nodes{}, Virtual: virtual, VirtualMap: make(map[string]string)}
}

func (r *Ring) Add(id string, host string, restPort int32, grpcPort int32) {
	r.Lock()
	defer r.Unlock()

	if r.Virtual == 0 {
		node := node.InitNode(id, host, restPort, grpcPort)
		r.Nodes = append(r.Nodes, node)
	} else {
		// Calculate the range for virtual nodes based on the number of virtual nodes
		virtualNodeRange := 1 << 31 / r.Virtual

		for i := 0; i < r.Virtual; i++ {
			// Calculate the virtual node ID within the range
			virtualNodeId := strconv.Itoa(int((hash(id) + uint32(virtualNodeRange*i)) % (1 << 31)))
			virtualId := id + "-" + virtualNodeId
			node := node.InitNode(virtualId, host, restPort, grpcPort)
			r.Nodes = append(r.Nodes, node)
			r.VirtualMap[virtualId] = id // map virtual node to actual node
		}
	}
	sort.Sort(r.Nodes)
}

func hash(s string) uint32 {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return binary.BigEndian.Uint32(bs[:4])
}

func (r *Ring) Remove(id string) error {
	r.Lock()
	defer r.Unlock()

	var newNodes node.Nodes
	var found bool

	if r.Virtual == 0 {
		// If no virtual nodes, simply remove the node with the matching ID
		i := r.search(id)
		if i >= r.Nodes.Len() || r.Nodes[i].Id != id {
			return errors.New("node not found")
		}

		r.Nodes = append(r.Nodes[:i], r.Nodes[i+1:]...)

		return nil
	} else {
		// Filter out nodes not associated with the removed node
		for _, n := range r.Nodes {
			if !strings.HasPrefix(n.Id, id+"-") {
				newNodes = append(newNodes, n)
			} else {
				delete(r.VirtualMap, n.Id) // Remove virtual node ID from map
				found = true               // Set found flag to true
			}
		}
	}

	// If the node was not found, return an error
	if !found {
		return errors.New("node not found")
	}

	r.Nodes = newNodes
	return nil
}

func (r *Ring) Get(id string) string {
	r.RLock()
	defer r.RUnlock()

	if len(r.Nodes) == 0 {
		panic("Empty ring")
	}
	if r.Virtual == 0 {
		i := r.search(id)
		if i >= r.Nodes.Len() {
			i = 0
		}

		return r.Nodes[i].Id
	} else {
		i := r.searchNode(id)
		if i >= r.Nodes.Len() {
			i = 0
		}
		return r.Nodes[i].Id
	}
}

func (r *Ring) search(id string) int {
	searchfn := func(i int) bool {
		return r.Nodes[i].HashId >= node.GetHashId(id)
	}

	return sort.Search(r.Nodes.Len(), searchfn)
}

func (r *Ring) searchNode(id string) int {
	hash := getHash(id)
	search := func(i int) bool {
		return r.Nodes[i].HashId >= hash
	}

	return sort.Search(r.Nodes.Len(), search)
}

func getHash(id string) uint32 {
	h := sha1.New()
	h.Write([]byte(id))
	bs := h.Sum(nil)
	return binary.BigEndian.Uint32(bs[:4])
}

func PrintBucketDistributionStats(filename string, buckets []string, members []string, maxVirtualNodes int) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	for virtualNodes := 1; virtualNodes <= maxVirtualNodes; virtualNodes++ {
		r := InitRing(virtualNodes)

		for _, bucket := range buckets {
			r.Add(bucket, "localhost", 8080, 5005)
		}

		statistics := r.GetDistributionStatistics(members)

		fmt.Fprintf(file, "No of virtual nodes : %d\n", virtualNodes)
		fmt.Fprintf(file, "%5s%10s\n", "%", "BucketId")
		for bucket, percent := range statistics {
			fmt.Fprintf(file, "%5.2f %s\n", percent, bucket)
		}
		fmt.Fprintln(file)
	}
}

func (r *Ring) GetDistributionStatistics(members []string) map[string]float64 {
	r.RLock()
	defer r.RUnlock()

	bucketCount := make(map[string]int)
	for _, member := range members {
		bucket := r.Get(member)
		bucketPrefix := strings.Split(bucket, "-")[0]
		bucketCount[bucketPrefix]++
	}

	statistics := make(map[string]float64)
	totalMembers := len(members)
	for bucketPrefix, count := range bucketCount {
		statistics[bucketPrefix] = (float64(count) / float64(totalMembers)) * 100
	}

	return statistics
}

func PrintBucketDistribution() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	buckets := make([]string, 10)
	for i := 0; i < 10; i++ {
		buckets[i] = strconv.Itoa(rand.Intn(1 << 31 / 5))
	}
	members := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		members[i] = strconv.Itoa(rand.Int())
	}
	maxVirtualNodes := 800
	PrintBucketDistributionStats("output.txt", buckets, members, maxVirtualNodes)
}
