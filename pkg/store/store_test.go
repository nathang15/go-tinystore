package store

import (
	"strconv"
	"testing"
	"time"
)

func TestSmallCache(t *testing.T) {
	// init an lru cache with capacity 2
	lru := Init(2)

	// test standard put and get
	lru.Put("2", "1")
	lru.Put("2", "2")
	actual, err := lru.Get("2")
	expected := "2"
	AssertEqualNoError(t, expected, actual, err)

	// test evict
	lru.Put("1", "1")
	lru.Put("4", "1")
	_, err = lru.Get("2")
	AssertErrorNoNil(t, err)
}

func TestLargeCache(t *testing.T) {
	// init an lru cache with capacity 200
	lru := Init(200)

	// test standard put and get
	lru.Put("1", "100")
	lru.Put("2", "200")
	actual, err := lru.Get("1")
	expected := "100"
	AssertEqualNoError(t, expected, actual, err)

	// test evict
	for i := 3; i <= 201; i++ {
		lru.Put(strconv.Itoa(i), strconv.Itoa(i*10))
	}
	_, err = lru.Get("2")
	AssertErrorNoNil(t, err)

	// test overwrite
	lru.Put("1", "500")
	actual, err = lru.Get("1")
	expected = "500"
	AssertEqualNoError(t, expected, actual, err)

	// test capacity limit
	for i := 202; i <= 400; i++ {
		lru.Put(strconv.Itoa(i), strconv.Itoa(i*10))
	}
	_, err = lru.Get("3")
	AssertErrorNoNil(t, err)
}

func AssertEqualNoError(t *testing.T, expected interface{}, actual interface{}, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if expected != actual {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
}

func AssertErrorNoNil(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Found element in cache when it should have been evicted")
	}
}

func TestCacheWriteThroughput(t *testing.T) {
	capacity := 100
	num_puts := 10000000
	lru := Init(capacity)
	start := time.Now()
	for i := 0; i < num_puts; i++ {
		v := strconv.Itoa(i)
		lru.Put(v, v)
	}
	end := time.Since(start)
	endSeconds := end.Seconds()
	t.Logf("Time to write 10M puts: %.2f seconds", endSeconds)
	t.Logf("LRU Cache write throughput: %.2f puts/second", float64(num_puts)/endSeconds)
}
