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
		for node.RNeighbor != nil && node.RNeighbor != m.Head {
			if(node.Status == ALIVE) {
				id := strconv.Itoa(node.Id)
				ts := node.Timestamp.String()
				member := id + ts
				list = append(list, member)
				node = node.RNeighbor
			}
		}

		if(node.Status == ALIVE) {
			id := strconv.Itoa(node.Id)
			ts := node.Timestamp.String()
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
		newNode.RNeighbor = m.Head.RNeighbor
		newNode.RrNeighbor = m.Head.RrNeighbor
		newNode.LNeighbor = m.Head
		newNode.LlNeighbor = m.Head.LNeighbor

		if(m.Head.LNeighbor != nil) {
			m.Head.LNeighbor.RrNeighbor = newNode
		}

		m.Head.RNeighbor = newNode
		m.Head.RrNeighbor = newNode.RNeighbor

		if(newNode.RNeighbor != nil) {
			newNode.RNeighbor.LNeighbor = newNode
			newNode.RNeighbor.LlNeighbor = m.Head
		}

		if(newNode.RrNeighbor != nil) {
			newNode.RrNeighbor.LlNeighbor = newNode
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
			m.Head = node.RNeighbor // TODO: check if rNeighbor is nil
		}

		node.RNeighbor.LNeighbor = node.LNeighbor
		node.RNeighbor.LlNeighbor = node.LlNeighbor
		node.LNeighbor.RrNeighbor = node.RrNeighbor
		node.LNeighbor.RNeighbor = node.RNeighbor
		node.LlNeighbor.RrNeighbor = node.RNeighbor
		node.RrNeighbor.LlNeighbor = node.LlNeighbor
		node.RrNeighbor = nil
		node.RNeighbor = nil
		node.LNeighbor = nil
		node.LlNeighbor = nil

		delete(m.NodeMap, id)
	}
	m.Mu.Unlock()
}