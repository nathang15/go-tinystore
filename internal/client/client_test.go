package client

import (
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	CONFIG_PATH     = "../../configs/nodes.json"
	CLIENT_CERT_DIR = "../../certs"
)

// ALL FOR DOCKER

// REST TESTS

func Test10kPutsNoVnode(t *testing.T) {
	vNode := 0
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	c := InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
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
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with REST API, no virtual node: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
}

func Test10kPuts(t *testing.T) {
	vNode := 10
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	c := InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
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
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with REST API, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
}

func Test50kPuts(t *testing.T) {
	vNode := 10
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	c := InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
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
	end := time.Since(start)
	t.Logf("Time to complete 50k puts with REST API, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/50,000 (%f%%)", int(miss), miss/50000)
}

// GRPC TESTS
func Test10kGRPCPutsNoVnode(t *testing.T) {
	vNode := 0
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	c := InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
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
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with GRPC, no virtual node: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
}
func Test10kGRPCPuts(t *testing.T) {
	vNode := 10
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	c := InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
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
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with GRPC, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)
}

func Test50kGRPCPuts(t *testing.T) {
	vNode := 10
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	c := InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
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
	end := time.Since(start)
	t.Logf("Time to complete 50k puts with GRPC, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/50,000 (%f%%)", int(miss), miss/50000)
}
