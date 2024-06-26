// LRU Cache implementation with doubly linked list and hashmap
package store

import (
	"errors"
	"sync"
)

type Node struct {
	prev *Node
	next *Node
	key  string
	val  string
}

type LRU struct {
	cache    map[string]*Node
	head     *Node
	tail     *Node
	capacity int
	size     int
	mut      sync.RWMutex // reader/writer mutual exclusion lock because this is read-heavy
}

func Init(capacity int) LRU {
	lru := LRU{
		cache:    make(map[string]*Node, capacity),
		head:     &Node{prev: nil, next: nil, key: "", val: ""},
		tail:     &Node{prev: nil, next: nil, key: "", val: ""},
		capacity: capacity,
		size:     0,
	}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

func (lru *LRU) Get(key string) (string, error) {
	lru.mut.RLock()
	defer lru.mut.RUnlock()
	if node, existed := lru.cache[key]; existed {
		lru.moveToHead(node)
		return node.val, nil
	}

	return "", errors.New("element not found")
}

func (lru *LRU) Put(key string, value string) {
	lru.mut.Lock()
	defer lru.mut.Unlock()
	if node, existed := lru.cache[key]; existed {
		node.val = value
		lru.moveToHead(node)
		return
	}

	node := Node{prev: lru.head, next: lru.head.next, key: key, val: value}

	lru.cache[key] = &node

	node.next = lru.head.next
	lru.head.next.prev = &node

	lru.head.next = &node
	node.prev = lru.head

	lru.size += 1

	// evict if exceed capacity
	if lru.size > lru.capacity {
		lru.evict()
		lru.size -= 1
	}
}

func (lru *LRU) moveToHead(node *Node) {
	// remove node from middle

	prev := node.prev
	next := node.next
	if prev != nil {
		prev.next = next
	}
	if next != nil {
		next.prev = prev
	}

	// add node to front

	node.next = lru.head.next
	lru.head.next.prev = node
	lru.head.next = node
	node.prev = lru.head
}

func (lru *LRU) evict() {
	node := lru.tail.prev
	prev := node.prev
	prev.next = lru.tail
	lru.tail.prev = prev

	delete(lru.cache, node.key)
}
