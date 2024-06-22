package main

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/nathang15/go-tinystore/internal/client"
	"github.com/nathang15/go-tinystore/internal/server"
)

const (
	CONFIG_PATH     = "configs/nodes-local.json"
	CLIENT_CERT_DIR = "certs"
)

func cleanupServers(t *testing.T, components []server.ServerConfig) {
	for _, srv_comps := range components {
		srv_comps.GrpcServer.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := srv_comps.HttpServer.Shutdown(ctx); err != nil {
			t.Logf("Http server shutdown error: %s", err)
		}
	}
}

func Test10kRestApiPuts(t *testing.T) {
	capacity := 100
	vNode := 0
	verbose := false
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	components := server.CreateAndRunAllFromConfig(capacity, configPath, verbose)

	c := client.InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 1000; i++ {
				v := strconv.Itoa(i)
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
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with REST API, 0 virtual node: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)

	cleanupServers(t, components)
}

func Test10kRestApiPuts10Vnode(t *testing.T) {
	capacity := 100
	vNode := 10
	verbose := false
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	components := server.CreateAndRunAllFromConfig(capacity, configPath, verbose)

	c := client.InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 1000; i++ {
				v := strconv.Itoa(i)
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
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with REST API, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)

	cleanupServers(t, components)
}

func Test50kRestApiPuts10Vnode(t *testing.T) {
	capacity := 100
	vNode := 10
	verbose := false
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	components := server.CreateAndRunAllFromConfig(capacity, configPath, verbose)

	c := client.InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 5000; i++ {
				v := strconv.Itoa(i)
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
	end := time.Since(start)
	t.Logf("Time to complete 50k puts with REST API, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/50,000 (%f%%)", int(miss), miss/50000)

	cleanupServers(t, components)
}

func Test10kGrpcPuts(t *testing.T) {
	capacity := 100
	verbose := false
	vNode := 0
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	components := server.CreateAndRunAllFromConfig(capacity, configPath, verbose)

	c := client.InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 1000; i++ {
				v := strconv.Itoa(i)
				err := c.PutForGrpc(v, v)
				if err != nil {
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with gRPC, no virtual node: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)

	cleanupServers(t, components)
}

func Test10kGrpcPuts10Vnode(t *testing.T) {
	capacity := 100
	verbose := false
	vNode := 10
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	components := server.CreateAndRunAllFromConfig(capacity, configPath, verbose)

	c := client.InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 1000; i++ {
				v := strconv.Itoa(i)
				err := c.PutForGrpc(v, v)
				if err != nil {
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	end := time.Since(start)
	t.Logf("Time to complete 10k puts with gRPC, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/10,000 (%f%%)", int(miss), miss/10000)

	cleanupServers(t, components)
}

func Test50kGrpcPuts10Vnode(t *testing.T) {
	capacity := 100
	verbose := false
	vNode := 10
	certdir, _ := filepath.Abs(CLIENT_CERT_DIR)
	configPath, _ := filepath.Abs(CONFIG_PATH)

	components := server.CreateAndRunAllFromConfig(capacity, configPath, verbose)

	c := client.InitClient(certdir, configPath, vNode)
	c.StartClusterConfigWatcher()

	var wg sync.WaitGroup
	var mutex sync.Mutex
	miss := 0.0

	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 5000; i++ {
				v := strconv.Itoa(i)
				err := c.PutForGrpc(v, v)
				if err != nil {
					mutex.Lock()
					miss += 1
					mutex.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	end := time.Since(start)
	t.Logf("Time to complete 50k puts with gRPC, 10 virtual nodes: %s", end)
	t.Logf("Cache misses: %d/50,000 (%f%%)", int(miss), miss/50000)

	cleanupServers(t, components)
}
