package membersList

import (
	"sync"
	"strconv"
	"fmt"
)

type MembersList struct {
	Head *Node
	NodeMap map[int]*Node // Map of node id to node pointer
	Mu sync.Mutex
}


func NewMembershipList() MembersList{
	return MembersList{ Head:nil, NodeMap:make(map[int]*Node) }
}

func (m *MembersList) GetHead() *Node {
	m.Mu.Lock()
	head := m.Head
	m.Mu.Unlock()
	return head
}

func (m *MembersList) GetNode(id int) *Node {
	m.Mu.Lock()
	node := m.NodeMap[id]
	m.Mu.Unlock()
	return node
}

func (m *MembersList) Read() [] string {
	var list []string

	m.Mu.Lock()
	node := m.Head
	if(node != nil) {
		for node.rNeighbor != nil && node.rNeighbor != m.Head {
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
	m.Mu.Unlock()

	return list
}

func (m *MembersList) Size() int {
	m.Mu.Lock()
	nMap := m.NodeMap
	m.Mu.Unlock()

	return len(nMap)
}

func (m *MembersList) Insert(newNode *Node) {
	fmt.Println("insert")
	id := newNode.GetId()
	m.Mu.Lock()
	if(m.Head == nil) {
		m.Head = newNode
	} else{
		newNode.rNeighbor = m.Head.rNeighbor
		newNode.rrNeighbor = m.Head.rrNeighbor
		newNode.lNeighbor = m.Head
		newNode.llNeighbor = m.Head.lNeighbor

		if(m.Head.lNeighbor != nil) {
			m.Head.lNeighbor.rrNeighbor = newNode
		}

		m.Head.rNeighbor = newNode
		m.Head.rrNeighbor = newNode.rNeighbor

		if(newNode.rNeighbor != nil) {
			newNode.rNeighbor.lNeighbor = newNode
			newNode.rNeighbor.llNeighbor = m.Head
		}

		if(newNode.rrNeighbor != nil) {
			newNode.rrNeighbor.llNeighbor = newNode
		}
	}

	m.NodeMap[id] = newNode
	m.Mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.Mu.Lock()
	node := m.NodeMap[id]
	
	if(node != nil) {
		// Choose new head if we remove current head
		if(node == m.Head) {
			m.Head = node.rNeighbor // TODO: check if rNeighbor is nil
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

		delete(m.NodeMap, id)
	}
	m.Mu.Unlock()
}