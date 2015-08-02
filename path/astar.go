package path

import (
	"container/heap"
)

type pqueue struct {
	a []*Item
	m map[Node]int
}
type Item struct {
	value    Node
	priority Weight
	index    int
}

func AStar(start, finish Node, heur Heuristic) Path {
	openlist := &pqueue{nil, make(map[Node]int)}
	closedlist := make(map[Node]bool)
	g := map[Node]Weight{start: 0}
	pre := map[Node]Node{}

	heap.Push(openlist, &Item{start, 0, 0})
	for len(openlist.a) > 0 {
		current := heap.Pop(openlist).(*Item).value
		if current == finish {
			return buildPath(start, finish, pre)
		}
		closedlist[current] = true
		expandNode(current, closedlist, openlist, g, pre, heur)
	}
	return nil
}

func expandNode(current Node, closed map[Node]bool, open *pqueue, g map[Node]Weight, pre map[Node]Node, heur Heuristic) {
	for _, conn := range current.Neighbors() {
		if closed[conn.To] {
			continue
		}
		tentative_g := g[current] + conn.W
		if _, ok := open.m[conn.To]; ok && tentative_g <= g[conn.To] {
			continue
		}
		pre[conn.To] = current
		g[conn.To] = tentative_g
		f := tentative_g + heur(current, conn.To)
		if _, ok := open.m[conn.To]; ok {
			open.a[open.m[conn.To]].priority = f
			heap.Fix(open, open.m[conn.To])
		} else {
			heap.Push(open, &Item{conn.To, f, 0})
		}
	}
}

func buildPath(start, finish Node, pre map[Node]Node) []Node {
	path := []Node{}

	for finish != start {
		path = append(path, finish)
		finish = pre[finish]
	}
	path = append(path, finish)

	// reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func (pq pqueue) Len() int { return len(pq.a) }

func (pq pqueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq.a[i].priority > pq.a[j].priority
}

func (pq pqueue) Swap(i, j int) {
	pq.a[i], pq.a[j] = pq.a[j], pq.a[i]
	pq.a[i].index = i
	pq.a[j].index = j
	pq.m[pq.a[i].value] = i
	pq.m[pq.a[j].value] = j
}

func (pq *pqueue) Push(x interface{}) {
	n := len(pq.a)
	item := x.(*Item)
	item.index = n
	pq.m[item.value] = n
	pq.a = append(pq.a, item)
}

func (pq *pqueue) Pop() interface{} {
	old := pq.a
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	pq.a = old[0 : n-1]
	delete(pq.m, item.value)
	return item
}
