package membersList

import (
	"time"
)

// Status of machine
const (
	ALIVE = iota	// 0
	JOIN			// 1
	LEAVE			// 2
	FAILED			// 3
)

type Node struct {
	id int
	hbCounter int
	timestamp time.Time
	rNeighbor *Node
	rrNeighbor *Node
	lNeighbor *Node
	llNeighbor *Node
	status int
}

func NewNode(id int, hbCount int, timestamp time.Time, rNeighbor *Node, rrNeighbor *Node, lNeighbor *Node, llNeighbor *Node, status int) * Node {
	return &Node{ id, hbCount, timestamp, rNeighbor, rrNeighbor, lNeighbor, llNeighbor, status }
}

func (n *Node) GetId() int {
	return n.id
}

func (n *Node) GetHBCount() int {
	return n.hbCounter
}

func (n *Node) GetTimestamp() time.Time {
	return n.timestamp
}

func (n *Node) GetNeighbors() (*Node, *Node, *Node, *Node) {
	return n.rNeighbor, n.rrNeighbor, n.lNeighbor, n.llNeighbor
}

func (n *Node) GetStatus() int {
	return n.status
}

func (n *Node) IncrementHBCounter() {
	n.hbCounter++
}

func (n *Node) SetHBCounter(hbCount int) {
	n.hbCounter = hbCount
}

func (n *Node) SetNeighbors(rNeighbor *Node, rrNeighbor *Node, lNeighbor *Node, llNeighbor *Node) {
	n.rNeighbor = rNeighbor
	n.rrNeighbor = rrNeighbor
	n.lNeighbor = lNeighbor
	n.llNeighbor = llNeighbor
}

func (n *Node) SetStatus(status int) {
	n.status = status
}

func (n *Node) Next() *Node {
	return n.rNeighbor
}
