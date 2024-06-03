package ring

import (
	"errors"
	"sort"
	"strconv"
	"sync"

	"github.com/nathang15/go-tinystore/node"
)

type Ring struct {
	Nodes   node.Nodes
	Virtual int
	sync.Mutex
}

func InitRing(virtual int) *Ring {
	return &Ring{Nodes: node.Nodes{}, Virtual: virtual}
}

func (r *Ring) Add(id string) {
	r.Lock()
	defer r.Unlock()

	for i := 0; i < r.Virtual; i++ {
		virtualId := id + "-" + strconv.Itoa(i)
		node := node.InitNode(virtualId)
		r.Nodes = append(r.Nodes, node)
	}

	sort.Sort(r.Nodes)
}

func (r *Ring) Remove(id string) error {
	r.Lock()
	defer r.Unlock()

	if len(r.Nodes) == 0 {
		return errors.New("node not found")
	}

	for i := 0; i < r.Virtual; i++ {
		virtualId := id + "-" + strconv.Itoa(i)
		err := r.remove(virtualId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Ring) remove(id string) error {
	i := r.search(id)
	if i >= r.Nodes.Len() || r.Nodes[i].Id != id {
		return errors.New("node not found")
	}

	r.Nodes = append(r.Nodes[:i], r.Nodes[i+1:]...)

	return nil
}

func (r *Ring) Get(id string) string {
	if r.Nodes.Len() == 0 {
		return ""
	}
	i := r.search(id)
	if i >= r.Nodes.Len() {
		i = 0
	}
	return r.Nodes[i].Id
}

func (r *Ring) search(id string) int {
	search := func(i int) bool {
		return r.Nodes[i].HashId >= node.GetHashId(id)
	}

	return sort.Search(r.Nodes.Len(), search)
}
