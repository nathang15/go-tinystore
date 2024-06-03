package ring

import (
	"fmt"
	"sort"
	"testing"

	"github.com/nathang15/go-tinystore/node"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	node1id = "node1"
	node2id = "node2"
	node3id = "node3"
	node4id = "node4"
	node5id = "node5"
)

func TestAddNode(t *testing.T) {
	Convey("Empty ring", t, func() {
		Convey("Add node", func() {
			r := InitRing()
			r.Add(node1id)
			So(r.Nodes.Len(), ShouldEqual, 1)
			Convey("Id should be hashed", func() {
				So(r.Nodes[0].HashId, ShouldHaveSameTypeAs, uint32(0))
			})
		})
		Convey("Add multiple nodes and sort by id", func() {
			r := InitRing()
			r.Add(node1id)
			r.Add(node2id)
			r.Add(node3id)
			So(r.Nodes.Len(), ShouldEqual, 3)
			node1hash := node.GetHashId(node1id)
			node2hash := node.GetHashId(node2id)
			node3hash := node.GetHashId(node3id)
			hashes := []uint32{node1hash, node2hash, node3hash}
			sort.Slice(hashes, func(i, j int) bool { return hashes[i] < hashes[j] })

			So(r.Nodes[0].HashId, ShouldEqual, hashes[0])
			So(r.Nodes[1].HashId, ShouldEqual, hashes[1])
			So(r.Nodes[2].HashId, ShouldEqual, hashes[2])
		})
	})
}

func TestRemoveNode(t *testing.T) {
	Convey("Existed nodes in ring", t, func() {
		r := InitRing()
		r.Add(node1id)
		r.Add(node2id)
		r.Add(node3id)
		Convey("Node not existed", func() {
			Convey("Return error", func() {
				err := r.Remove("NotExistedNode")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node not found")
			})
		})
		Convey("Node exists", func() {
			Convey("Remove node", func() {
				err := r.Remove(node2id)
				So(err, ShouldBeNil)
				So(r.Nodes.Len(), ShouldEqual, 2)
				node3hash := node.GetHashId(node3id)
				node1hash := node.GetHashId(node1id)
				hashes := []uint32{node3hash, node1hash}
				sort.Slice(hashes, func(i, j int) bool { return hashes[i] < hashes[j] })

				So(r.Nodes[0].HashId, ShouldEqual, hashes[0])
				So(r.Nodes[1].HashId, ShouldEqual, hashes[1])
			})
		})
	})
	Convey("Remove from empty ring", t, func() {
		r := InitRing()
		err := r.Remove(node1id)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "node not found")
	})
	Convey("Remove all nodes", t, func() {
		r := InitRing()
		r.Add(node1id)
		r.Add(node2id)
		err1 := r.Remove(node1id)
		err2 := r.Remove(node2id)
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)
		So(r.Nodes.Len(), ShouldEqual, 0)
	})
}

func TestGet(t *testing.T) {
	Convey("Ring with 1 node", t, func() {
		r := InitRing()
		r.Add(node1id)
		Convey("Return that node regardless of input", func() {
			insertnode := r.Get("id")
			So(insertnode, ShouldEqual, node1id)
			insertnode = r.Get("anykey")
			So(insertnode, ShouldEqual, node1id)
		})
	})

	Convey("Ring with multiple nodes", t, func() {
		r := InitRing()
		r.Add(node1id)
		r.Add(node2id)
		r.Add(node3id)
		Convey("Return closest node", func() {
			insertid := "justa"
			insertHash := node.GetHashId(insertid)

			nodes := r.Nodes
			nodeHashes := []uint32{
				node.GetHashId(node1id),
				node.GetHashId(node2id),
				node.GetHashId(node3id),
			}
			sort.Slice(nodeHashes, func(i, j int) bool { return nodeHashes[i] < nodeHashes[j] })

			var expectedNode string
			if insertHash <= nodeHashes[0] || insertHash > nodeHashes[len(nodeHashes)-1] {
				expectedNode = nodes[0].Id // wrap around to the first node
			} else {
				for i, hash := range nodeHashes {
					if insertHash <= hash {
						expectedNode = nodes[i].Id
						break
					}
				}
			}
			insertnode := r.Get(insertid)
			So(insertnode, ShouldEqual, expectedNode)
		})
	})

	Convey("Edge case: Node hashes wrap around", t, func() {
		r := InitRing()
		r.Add(node1id)
		r.Add(node2id)
		r.Add(node3id)
		r.Add(node4id)
		r.Add(node5id)

		Convey("Return closest node with wrap around", func() {
			insertid := "wraparound"
			insertHash := node.GetHashId(insertid)

			nodes := r.Nodes
			nodeHashes := []uint32{
				node.GetHashId(node1id),
				node.GetHashId(node2id),
				node.GetHashId(node3id),
				node.GetHashId(node4id),
				node.GetHashId(node5id),
			}
			sort.Slice(nodeHashes, func(i, j int) bool { return nodeHashes[i] < nodeHashes[j] })

			var expectedNode string
			if insertHash <= nodeHashes[0] || insertHash > nodeHashes[len(nodeHashes)-1] {
				expectedNode = nodes[0].Id // wrap around to the first node
			} else {
				for i, hash := range nodeHashes {
					if insertHash <= hash {
						expectedNode = nodes[i].Id
						break
					}
				}
			}
			insertnode := r.Get(insertid)
			So(insertnode, ShouldEqual, expectedNode)
		})
	})

	Convey("Empty ring", t, func() {
		r := InitRing()
		insertnode := r.Get("anykey")
		So(insertnode, ShouldBeEmpty)
	})
}

func TestStress(t *testing.T) {
	Convey("Stress test with many nodes", t, func() {
		r := InitRing()
		for i := 0; i < 10000; i++ {
			r.Add(fmt.Sprintf("node%d", i))
		}
		So(r.Nodes.Len(), ShouldEqual, 10000)
		Convey("Remove half the nodes", func() {
			for i := 0; i < 5000; i++ {
				r.Remove(fmt.Sprintf("node%d", i))
			}
			So(r.Nodes.Len(), ShouldEqual, 5000)
		})
	})
}
