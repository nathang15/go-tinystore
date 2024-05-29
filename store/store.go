package store

type Node struct {
	prev *Node
	next *Node
	key  int
	val  int
}

type LRU struct {
	cache    map[int]*Node
	head     *Node
	tail     *Node
	capacity int
	size     int
}

func Init(capacity int) LRU {
	lru := LRU{
		cache:    make(map[int]*Node, capacity),
		head:     &Node{prev: nil, next: nil, key: -1, val: -1},
		tail:     &Node{prev: nil, next: nil, key: -1, val: -1},
		capacity: capacity,
		size:     0,
	}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

func (lru *LRU) Get(key int) int {

	if node, existed := lru.cache[key]; existed {
		lru.moveToHead(node)
		return node.val
	}

	return -1
}

func (lru *LRU) Put(key int, value int) {

	if node, existed := lru.cache[key]; existed {
		node.val = value
		lru.moveToHead(node)
		return
	}

	node := Node{prev: nil, next: nil, key: key, val: value}

	lru.cache[key] = &node

	node.next = lru.head.next
	lru.head.next.prev = &node
	lru.size += 1

	// evict if exceed capacity
	if lru.size > lru.capacity {
		dNode := lru.tail.prev
		prev := dNode.prev
		prev.next = lru.tail
		lru.tail.prev = prev

		delete(lru.cache, dNode.key)

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
