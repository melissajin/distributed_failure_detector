package membersList

import (
	"sync"
	"strconv"
	"fmt"
)

type MembersList struct {
	Head *Node
	NodeMap map[int]*Node // Map of node id to node pointer
	mu sync.Mutex
}


func NewMembershipList() MembersList{
	return MembersList{ Head:nil, NodeMap:make(map[int]*Node) }
}

func (m *MembersList) GetHead() *Node {
	m.mu.Lock()
	head := m.Head
	m.mu.Unlock()
	return head
}

func (m *MembersList) GetNode(id int) *Node {
	m.mu.Lock()
	node := m.NodeMap[id]
	m.mu.Unlock()
	return node
}

func (m *MembersList) Read() [] string {
	var list []string

	m.mu.Lock()
	node := m.Head
	for node != nil {
		if(node.Status == ALIVE) {
			id := strconv.Itoa(node.Id)
			ts := node.Timestamp
			member := "Machine Id: " + id + " Timestamp: " + ts
			list = append(list, member)
		}
		node = node.Next()
		if node == m.Head {
			break
		}
	}
	m.mu.Unlock()

	return list
}

func (m *MembersList) Size() int {
	m.mu.Lock()
	nMap := m.NodeMap
	m.mu.Unlock()

	return len(nMap)
}

func (m *MembersList) Insert(newNode *Node) {
	id := newNode.GetId()
	m.mu.Lock()
	fmt.Println(m.Head)
	if(m.Head == nil) {
		m.Head = newNode
		newNode.Right = m.Head
		newNode.Left = m.Head
	} else{
		newNode.Left = m.Head.Left
		newNode.Right = m.Head

		m.Head.Left.Right = newNode
		m.Head.Left = newNode
	}
	fmt.Println(m.Head)
	m.NodeMap[id] = newNode
	m.mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
	node := m.NodeMap[id]
	
	if(node != nil) {
		// Choose new head if we remove current head
		if(node == m.Head) {
			m.Head = m.Head.Right
		}

		node.Left.Right = node.Right
		node.Right.Left = node.Left

		node.Left = nil
		node.Right = nil

		delete(m.NodeMap, id)
	}
	m.mu.Unlock()
}