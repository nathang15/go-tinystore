package ring

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nathang15/go-tinystore/node"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddNode(t *testing.T) {
	Convey("Add node to ring with 1 virtual node", t, func() {
		r := InitRing(1)
		r.Add("node1", "localhost", 8080)

		Convey("Node should be added", func() {
			So(r.Nodes.Len(), ShouldEqual, 1)
		})

		Convey("Id should be hashed", func() {
			So(r.Nodes[0].HashId, ShouldHaveSameTypeAs, uint32(0))
		})
	})

	Convey("Add node to ring with 5 virtual nodes", t, func() {
		r := InitRing(5)
		r.Add("node1", "localhost", 8080)

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
		r.Add("node1", "localhost", 8080)
		r.Add("node2", "localhost", 8081)
		r.Add("node3", "localhost", 8082)

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
		r.Add("node1", "localhost", 8080)
		r.Add("node2", "localhost", 8081)
		r.Add("node3", "localhost", 8082)

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
		r.Add("node1", "localhost", 8080)
		r.Add("node2", "localhost", 8081)
		r.Add("node3", "localhost", 8082)

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

func getPrefix(s string) string {
	parts := strings.SplitN(s, "-", 2)
	return parts[0]
}

func TestGet(t *testing.T) {
	// Load the nodes configuration from the JSON file
	nodes_config := node.LoadNodesConfig("../configs/nodes.json")
	node0 := nodes_config.Nodes["node0"]
	node1 := nodes_config.Nodes["node1"]
	node2 := nodes_config.Nodes["node2"]

	Convey("Given a ring with 1 vnode", t, func() {
		r := InitRing(1)
		r.Add(node1.Id, node1.Host, node1.Port)

		Convey("Then it should return that node regardless of input", func() {
			insertnode := r.Get("id")
			So(getPrefix(insertnode), ShouldEqual, getPrefix(node1.Id))

			insertnode = r.Get("anykey")
			So(getPrefix(insertnode), ShouldEqual, getPrefix(node1.Id))
		})
	})

	Convey("Given a ring with multiple vnodes", t, func() {
		insertid := "random_key"

		r := InitRing(3)
		r.Add(node0.Id, node0.Host, node0.Port)
		r.Add(node1.Id, node1.Host, node1.Port)
		r.Add(node2.Id, node2.Host, node2.Port)

		Convey("Then it should return the node closest to the hashed key", func() {
			node0hash := node.GetHashId(node0.Id)
			node1hash := node.GetHashId(node1.Id)
			inserthash := node.GetHashId(insertid)

			So(inserthash, ShouldBeGreaterThanOrEqualTo, node1hash)
			So(inserthash, ShouldBeLessThan, node0hash)

			insertnode := r.Get(insertid)
			So(getPrefix(insertnode), ShouldEqual, getPrefix(node0.Id))
		})
	})
}

func TestStress(t *testing.T) {
	Convey("Stress test with many nodes and virtual nodes", t, func() {
		r := InitRing(100)
		for i := 0; i < 1000; i++ {
			r.Add(fmt.Sprintf("node%d", i), "localhost", int32(8080+i))
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
		r.Add("node1", "localhost", 8080)

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
			r.Add("node1", "localhost", 8080)
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
