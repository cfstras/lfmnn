package path

type Weight float32

type Graph interface {
	Nodes() []Node
}

type Node interface {
	String() string
	// Neighbors returns an array of Connections
	Neighbors() []Connection
	// Whether neighbors are sorted by path weight
	IsSorted() bool
}

type Connection struct {
	From, To Node
	W        Weight
}

type Heuristic func(Node, Node) Weight

type Path []Node
