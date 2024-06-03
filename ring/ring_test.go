package ring

import (
	"fmt"
	"sort"
	"testing"

	"github.com/nathang15/go-tinystore/node"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddNode(t *testing.T) {
	Convey("Add node to ring with 1 virtual node", t, func() {
		r := InitRing(1)
		r.Add("node1")

		Convey("Node should be added", func() {
			So(r.Nodes.Len(), ShouldEqual, 1)
		})

		Convey("Id should be hashed", func() {
			So(r.Nodes[0].HashId, ShouldHaveSameTypeAs, uint32(0))
		})
	})

	Convey("Add node to ring with 5 virtual nodes", t, func() {
		r := InitRing(5)
		r.Add("node1")

		Convey("5 virtual nodes should be added", func() {
			So(r.Nodes.Len(), ShouldEqual, 5)
		})

		Convey("Ids should be hashed", func() {
			for i := 0; i < 5; i++ {
				So(r.Nodes[i].HashId, ShouldHaveSameTypeAs, uint32(0))
			}
		})
	})

	Convey("Add multiple nodes to ring with 3 virtual nodes", t, func() {
		r := InitRing(3)
		r.Add("node1")
		r.Add("node2")
		r.Add("node3")

		Convey("9 virtual nodes should be added", func() {
			So(r.Nodes.Len(), ShouldEqual, 9)
		})

		Convey("Ids should be hashed", func() {
			for i := 0; i < 9; i++ {
				So(r.Nodes[i].HashId, ShouldHaveSameTypeAs, uint32(0))
			}
		})
	})
}

func TestRemoveNode(t *testing.T) {
	Convey("Remove node from ring with 1 virtual node", t, func() {
		r := InitRing(1)
		r.Add("node1")
		r.Add("node2")
		r.Add("node3")

		Convey("Node should be removed", func() {
			err := r.Remove("node2")
			So(err, ShouldBeNil)
			So(r.Nodes.Len(), ShouldEqual, 2)
		})

		Convey("Error when node not found", func() {
			err := r.Remove("node4")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node not found")
		})
	})

	Convey("Remove node from ring with 5 virtual nodes", t, func() {
		r := InitRing(5)
		r.Add("node1")
		r.Add("node2")
		r.Add("node3")

		Convey("5 virtual nodes should be removed", func() {
			err := r.Remove("node2")
			So(err, ShouldBeNil)
			So(r.Nodes.Len(), ShouldEqual, 10)
		})

		Convey("Error when node not found", func() {
			err := r.Remove("node4")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node not found")
		})
	})
}

func TestGet(t *testing.T) {
	Convey("Get node from ring with 1 virtual node", t, func() {
		r := InitRing(1)
		r.Add("node1")
		r.Add("node2")
		r.Add("node3")

		Convey("Return closest node", func() {
			insertid := "justa"
			insertHash := node.GetHashId(insertid)

			nodes := r.Nodes
			nodeHashes := []uint32{
				node.GetHashId("node1-0"),
				node.GetHashId("node2-0"),
				node.GetHashId("node3-0"),
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

	Convey("Get node from ring with 5 virtual nodes", t, func() {
		r := InitRing(5)
		r.Add("node1")
		r.Add("node2")
		r.Add("node3")

		Convey("Return closest node", func() {
			insertid := "justa"
			insertHash := node.GetHashId(insertid)

			nodes := r.Nodes
			nodeHashes := []uint32{
				node.GetHashId("node1-0"),
				node.GetHashId("node1-1"),
				node.GetHashId("node1-2"),
				node.GetHashId("node1-3"),
				node.GetHashId("node1-4"),
				node.GetHashId("node2-0"),
				node.GetHashId("node2-1"),
				node.GetHashId("node2-2"),
				node.GetHashId("node2-3"),
				node.GetHashId("node2-4"),
				node.GetHashId("node3-0"),
				node.GetHashId("node3-1"),
				node.GetHashId("node3-2"),
				node.GetHashId("node3-3"),
				node.GetHashId("node3-4"),
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
}

func TestStress(t *testing.T) {
	Convey("Stress test with many nodes and virtual nodes", t, func() {
		r := InitRing(100)
		for i := 0; i < 1000; i++ {
			r.Add(fmt.Sprintf("node%d", i))
		}
		So(r.Nodes.Len(), ShouldEqual, 100000)
		Convey("Remove half the nodes", func() {
			for i := 0; i < 500; i++ {
				err := r.Remove(fmt.Sprintf("node%d", i))
				So(err, ShouldBeNil)
			}
			So(r.Nodes.Len(), ShouldEqual, 50000)
		})
	})
}

func TestEdgeCases(t *testing.T) {
	Convey("Edge cases", t, func() {
		r := InitRing(1)
		r.Add("node1")

		Convey("Get node with empty ring", func() {
			r := InitRing(1)
			insertnode := r.Get("anykey")
			So(insertnode, ShouldBeEmpty)
		})

		Convey("Remove node from empty ring", func() {
			r := InitRing(1)
			err := r.Remove("node1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node not found")
		})

		Convey("Add node to ring with 0 virtual nodes", func() {
			r := InitRing(0)
			r.Add("node1")
			So(r.Nodes.Len(), ShouldEqual, 0)
		})

		Convey("Remove node from ring with 0 virtual nodes", func() {
			r := InitRing(0)
			err := r.Remove("node1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node not found")
		})
	})
}
