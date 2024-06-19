package client

import (
	"strconv"
	"sync"
	"testing"
)

// ALL FOR DOCKER

// REST BENCHMARKS
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
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	t.Logf("Cache misses: %d/100,000 (%f%%)", int(miss), miss/100000)
}

// func Benchmark50kConcurrentPuts(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 50; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 1000; j++ {
// 					v := strconv.Itoa(j)
// 					c.Put(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func Benchmark100kConcurrentPuts(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 100; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 1000; j++ {
// 					v := strconv.Itoa(j)
// 					c.Put(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func BenchmarkLargeVolumeConcurrentPuts(b *testing.B) {
// 	largeValue := make([]byte, 1024*1024) // 1MB value
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 10; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 100; j++ {
// 					v := strconv.Itoa(j)
// 					c.Put(v, string(largeValue))
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func BenchmarkHighIterationConcurrentPuts(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 10; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 10000; j++ {
// 					v := strconv.Itoa(j)
// 					c.Put(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// // GRPC BENCHMARKS
// func Benchmark10kConcurrentPutsGrpc(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 10; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 1000; j++ {
// 					v := strconv.Itoa(j)
// 					c.PutForGrpc(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func Benchmark50kConcurrentPutsGrpc(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 50; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 1000; j++ {
// 					v := strconv.Itoa(j)
// 					c.PutForGrpc(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func Benchmark100kConcurrentPutsGrpc(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 100; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 1000; j++ {
// 					v := strconv.Itoa(j)
// 					c.PutForGrpc(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func BenchmarkLargeVolumeConcurrentPutsGrpc(b *testing.B) {
// 	largeValue := make([]byte, 1024*1024) // 1MB value
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 10; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 100; j++ {
// 					v := strconv.Itoa(j)
// 					c.PutForGrpc(v, string(largeValue))
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func BenchmarkHighIterationConcurrentPutsGrpc(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		c := InitClient("../configs/nodes.json", 100)
// 		var wg sync.WaitGroup
// 		for i := 0; i < 10; i++ {
// 			wg.Add(1)
// 			go func() {
// 				defer wg.Done()
// 				for j := 1; j <= 10000; j++ {
// 					v := strconv.Itoa(j)
// 					c.PutForGrpc(v, v)
// 				}
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }
