package membersList

import (
	"time"
)

// Status of machine
const (
	ALIVE = iota	// 0
	LEAVE			// 1
	FAILED			// 2
)

type Node struct {
	Id int
	HbCounter int
	Timestamp string
	RNeighbor *Node
	RrNeighbor *Node
	LNeighbor *Node
	LlNeighbor *Node
	Status int
}

func NewNode(id int, hbCount int, timestamp string, rNeighbor *Node, rrNeighbor *Node, lNeighbor *Node, llNeighbor *Node, status int) * Node {
	return &Node{ id, hbCount, timestamp, rNeighbor, rrNeighbor, lNeighbor, llNeighbor, status }
}

func (n *Node) GetId() int {
	return n.Id
}

func (n *Node) GetHBCount() int {
	return n.HbCounter
}

func (n *Node) GetTimestamp() string {
	return n.Timestamp
}

func (n *Node) GetNeighbors() (*Node, *Node, *Node, *Node) {
	return n.RNeighbor, n.RrNeighbor, n.LNeighbor, n.LlNeighbor
}

func (n *Node) GetStatus() int {
	return n.Status
}

func (n *Node) IncrementHBCounter() {
	n.HbCounter++
}

func (n *Node) SetHBCounter(hbCount int) {
	n.HbCounter = hbCount
}

func (n *Node) SetNeighbors(rNeighbor *Node, rrNeighbor *Node, lNeighbor *Node, llNeighbor *Node) {
	n.RNeighbor = rNeighbor
	n.RrNeighbor = rrNeighbor
	n.LNeighbor = lNeighbor
	n.LlNeighbor = llNeighbor
}

func (n *Node) SetStatus(status int) {
	n.Status = status
}

func (n *Node) Next() *Node {
	return n.RNeighbor
}
