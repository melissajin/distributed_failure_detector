package node

import (
	"time"
	"sync"
	"strconv"
)

// Status of machine, sent in heartbeat
const (
	ALIVE = iota	// 0
	JOIN			// 1
	LEAVE			// 2
	FAILED			// 3
)

type Heartbeat struct {
	id int
	host string
	membershipList MembersList
	timestamp time.Time
	status int
}

type MembersList struct {
	head *Node
	mu sync.Mutex
}

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

func (m *MembersList) Head() *Node{
	m.mu.Lock()
	node := m.head
	m.mu.Unlock()
	return node
}

func (hb *Heartbeat) GetInfo() (int, string, MembersList, time.Time, int) {
	return hb.id, hb.host, hb.membershipList, hb.timestamp, hb.status
}

func (hb *Heartbeat) SetInfo(id int, host string, membershipList MembersList, timestamp time.Time, status int) {
	hb.id = id
	hb.host = host
	hb.membershipList = membershipList
	hb.timestamp = timestamp
	hb.status = status
}

func (n *Node) GetInfo() (int, int, time.Time, *Node, *Node, *Node, *Node, int) {
	return n.id, n.hbCounter, n.timestamp, n.rNeighbor, n.rrNeighbor, n.lNeighbor, n.llNeighbor, n.status
}

func (n *Node) SetInfo(hbCount int, timestamp time.Time, rNeighbor *Node, rrNeighbor *Node, lNeighbor *Node, llNeighbor *Node, status int) {
	n.hbCounter = hbCount
	n.timestamp = timestamp
	n.rNeighbor = rNeighbor
	n.rrNeighbor = rrNeighbor
	n.lNeighbor = lNeighbor
	n.llNeighbor = llNeighbor
	n.status = status
}

func (n *Node) GetHbCounter() int {
	return n.hbCounter
}

func (n *Node) SetHbCounter(hbCount int) {
	n.hbCounter = hbCount
}

func (m *MembersList) Read() [] string{
	m.mu.Lock()
	var list []string
	node := m.head
	for node.rNeighbor != m.head {
		if(node.status == ALIVE) {
			id := strconv.Itoa(node.id)
			ts := node.timestamp.String()
			member := id + ts
			list = append(list, member)
			node = node.rNeighbor
		}
	}
	if(node.status == ALIVE) {
		id := strconv.Itoa(node.id)
		ts := node.timestamp.String()
		member := id + ts
		list = append(list, member)
	}
	m.mu.Unlock()
	return list
}

func (m *MembersList) Length() int {
	m.mu.Lock()
	var length int
	if(m.head == nil) {
		length = 0
	} else {
		length = len(m.Read())
	}
	m.mu.Unlock()
	return length
}

func (m *MembersList) Insert(id int, timestamp time.Time, hbCounter int) {
	m.mu.Lock()
	newNode := &Node{id, hbCounter, timestamp, nil, nil, nil, nil, ALIVE}
	newNode.rNeighbor = m.head.rNeighbor
	newNode.rrNeighbor = m.head.rrNeighbor
	newNode.lNeighbor = m.head
	newNode.llNeighbor = m.head.lNeighbor
	m.head.rNeighbor = newNode
	m.head.rrNeighbor = newNode.rNeighbor
	m.head.lNeighbor.rNeighbor = newNode
	newNode.rNeighbor.lNeighbor = newNode
	newNode.rNeighbor.llNeighbor = m.head
	newNode.rrNeighbor. llNeighbor = newNode
	IdMap[id] = newNode
	m.mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
	node := m.head
	for (node.id != id) {
		node = node.rNeighbor
	}
	if(node == m.head) {
		m.head = node.rNeighbor
	}
	node.rNeighbor.lNeighbor = node.lNeighbor
	node.rNeighbor.llNeighbor = node.llNeighbor
	node.lNeighbor.rrNeighbor = node.rrNeighbor
	node.lNeighbor.rNeighbor = node.rNeighbor
	node.llNeighbor.rrNeighbor = node.rNeighbor
	node.rrNeighbor.llNeighbor = node.llNeighbor
	node.rrNeighbor = nil
	node.rNeighbor = nil
	node.lNeighbor = nil
	node.llNeighbor = nil
	delete(IdMap, id)
	m.mu.Unlock()
}

var IdMap map[int]*Node
