package client

import (
	"strconv"
	"sync"
	"testing"
)

// ALL FOR DOCKER

// REST TESTS
func Test10kConcurrentPuts(t *testing.T) {
	c := InitClient("../configs/nodes.json", 10)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 1000; j++ {
				v := strconv.Itoa(j)
				err := c.Put(v, v)
				if err != nil {
					t.Logf("Error putting key %s: %v", v, err)
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
}

func Test50kConcurrentPuts(t *testing.T) {
	c := InitClient("../configs/nodes.json", 10)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 1000; j++ {
				v := strconv.Itoa(j)
				err := c.Put(v, v)
				if err != nil {
					t.Logf("Error putting key %s: %v", v, err)
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	t.Logf("Cache misses: %d/50,000 (%f%%)", int(miss), miss/50000)
}

// func TestLargeVolumeConcurrentPuts(t *testing.T) {
// 	largeValue := make([]byte, 1024*1024) // 1MB value
// 	c := InitClient("../configs/nodes.json", 100)
// 	c.StartClusterConfigWatcher()
// 	var wg sync.WaitGroup
// 	var mutex sync.Mutex
// 	miss := 0.0
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for j := 1; j <= 1000; j++ {
// 				v := strconv.Itoa(j)
// 				err := c.Put(v, string(largeValue))
// 				if err != nil {
// 					t.Logf("Error putting key %s: %v", v, err)
// 					mutex.Lock()
// 					miss += 1
// 					mutex.Unlock()
// 				}
// 			}
// 		}()
// 	}
// 	wg.Wait()
// 	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
// }

// GRPC TESTS
func Test10kConcurrentGRPCPuts(t *testing.T) {
	c := InitClient("../configs/nodes.json", 100)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 1000; j++ {
				v := strconv.Itoa(j)
				err := c.PutForGrpc(v, v)
				if err != nil {
					t.Logf("Error putting key %s: %v", v, err)
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
}

func Test50kConcurrentGRPCPuts(t *testing.T) {
	c := InitClient("../configs/nodes.json", 100)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= 1000; j++ {
				v := strconv.Itoa(j)
				err := c.PutForGrpc(v, v)
				if err != nil {
					t.Logf("Error putting key %s: %v", v, err)
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	t.Logf("Cache misses: %d/50,000 (%f%%)", int(miss), miss/50000)
}

// func TestLargeVolumeConcurrentGRPCPuts(t *testing.T) {
// 	largeValue := make([]byte, 1024*1024) // 1MB value
// 	c := InitClient("../configs/nodes.json", 100)
// 	c.StartClusterConfigWatcher()
// 	var wg sync.WaitGroup
// 	var mutex sync.Mutex
// 	miss := 0.0
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for j := 1; j <= 1000; j++ {
// 				v := strconv.Itoa(j)
// 				err := c.PutForGrpc(v, string(largeValue))
// 				if err != nil {
// 					t.Logf("Error putting key %s: %v", v, err)
// 					mutex.Lock()
// 					miss += 1
// 					mutex.Unlock()
// 				}
// 			}
// 		}()
// 	}
// 	wg.Wait()
// 	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
// }
