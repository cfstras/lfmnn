package path_test

import (
	. "."
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type suite struct{}

var _ = Suite(&suite{})

type graph struct {
	nodes []node
}

type node struct {
	parent    *graph
	name      string
	neighbors []Connection
}

var graph1 graph = graph{}

func init() {
	graph1.nodes = []node{
		node{&graph1, "0", []Connection{}},
		node{&graph1, "A", make([]Connection, 1)},
		node{&graph1, "B", make([]Connection, 1)},
		node{&graph1, "C", nil},
	}
	graph1.nodes[1].neighbors[0] = Connection{&graph1.nodes[1], &graph1.nodes[2], 1}
	graph1.nodes[2].neighbors[0] = Connection{&graph1.nodes[2], &graph1.nodes[3], 1}
}

func (node) IsSorted() bool {
	return false
}

func (n *node) Neighbors() []Connection {
	return n.neighbors
}

func (n node) String() string {
	return n.name
}

func (s *suite) TestAStarSimple(c *C) {
	expected := Path{&graph1.nodes[0]}
	got := AStar(&graph1.nodes[0], &graph1.nodes[0], func(Node, Node) Weight { return 0 })
	c.Assert(got, DeepEquals, expected)

	expected = Path{&graph1.nodes[1], &graph1.nodes[2], &graph1.nodes[3]}
	got = AStar(&graph1.nodes[1], &graph1.nodes[3], func(Node, Node) Weight { return 0 })
	c.Assert(got, DeepEquals, expected)
}
