package membersList

import (
	"fmt"
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
	Right *Node
	Left *Node
	Status int
}

func NewNode(id int, hbCount int, timestamp string, status int) * Node {
	fmt.Println("CHECK")
	return &Node{ id, hbCount, timestamp, nil, nil, status }
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
	r := n.Right
	l := n.Left

	var rr *Node
	var ll *Node
	rr = nil
	ll = nil

	if r != nil {
		rr = r.Right
	}
	if l != nil {
		ll = l.Left
	}
	return r, rr, l, ll
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

func (n *Node) SetNeighbors(right *Node, left *Node) {
	n.Right = right
	n.Left = left
}

func (n *Node) SetStatus(status int) {
	n.Status = status
}

func (n *Node) Next() *Node {
	return n.Right
}
