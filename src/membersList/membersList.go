package membersList

import (
	"sync"
	"strconv"
	// "fmt"
)

type MembersList struct {
	head *Node
	mu sync.Mutex
}

// Map of node id to node pointer
var nodeMap map[int]*Node

func NewMembershipList(head *Node) *MembersList{
	nodeMap[head.GetId()] = head
	return &MembersList{ head:head }
}

func (m *MembersList) GetHead() *Node {
	m.mu.Lock()
	node := m.head
	m.mu.Unlock()
	return node
}

func (m *MembersList) GetNode(id int) *Node {
	return nodeMap[id]
}

func (m *MembersList) Read() [] string {
	var list []string

	m.mu.Lock()
	node := m.head
	if(node != nil) {
		for node.rNeighbor != nil && node.rNeighbor != m.head {
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
	}
	m.mu.Unlock()

	return list
}

// func (m *MembersList) Size() int {
// 	m.mu.Lock()
// 	head := m.head
// 	m.mu.Unlock()

// 	if(head == nil) {
// 		return 0
// 	} else {
// 		return len(m.Read())
// 	}
// }

func (m *MembersList) Insert(newNode *Node) {
	id := newNode.GetId()
	m.mu.Lock()
	if(m.head == nil) {
		m.head = newNode
	} else{
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
		nodeMap[id] = newNode
	}
	m.mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
	node := nodeMap[id]
	
	if(node != nil) {
		// Choose new head if we remove current head
		if(node == m.head) {
			m.head = node.rNeighbor // TODO: check if rNeighbor is nil
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

		delete(nodeMap, id)
	}
	m.mu.Unlock()
}