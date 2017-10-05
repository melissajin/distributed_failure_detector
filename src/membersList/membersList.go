package membersList

import (
	"sync"
	"strconv"
	"fmt"
)

type MembersList struct {
	head *Node
	nodeMap map[int]*Node // Map of node id to node pointer
	mu sync.Mutex
}


func NewMembershipList() MembersList{
	return MembersList{ head:nil, nodeMap:make(map[int]*Node) }
}

func (m *MembersList) GetHead() *Node {
	m.mu.Lock()
	head := m.head
	m.mu.Unlock()
	return head
}

func (m *MembersList) GetNode(id int) *Node {
	m.mu.Lock()
	node := m.nodeMap[id]
	m.mu.Unlock()
	return node
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

func (m *MembersList) Size() int {
	m.mu.Lock()
	nMap := m.nodeMap
	m.mu.Unlock()

	return len(nMap)
}

func (m *MembersList) Insert(newNode *Node) {
	fmt.Println("insert")
	id := newNode.GetId()
	m.mu.Lock()
	if(m.head == nil) {
		m.head = newNode
	} else{
		newNode.rNeighbor = m.head.rNeighbor
		newNode.rrNeighbor = m.head.rrNeighbor
		newNode.lNeighbor = m.head
		newNode.llNeighbor = m.head.lNeighbor

		if(m.head.lNeighbor != nil) {
			m.head.lNeighbor.rrNeighbor = newNode
		}

		m.head.rNeighbor = newNode
		m.head.rrNeighbor = newNode.rNeighbor

		if(newNode.rNeighbor != nil) {
			newNode.rNeighbor.lNeighbor = newNode
			newNode.rNeighbor.llNeighbor = m.head
		}

		if(newNode.rrNeighbor != nil) {
			newNode.rrNeighbor.llNeighbor = newNode
		}
	}

	m.nodeMap[id] = newNode
	m.mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
	node := m.nodeMap[id]
	
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

		delete(m.nodeMap, id)
	}
	m.mu.Unlock()
}