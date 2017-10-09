package membersList

import (
	"sync"
)

const (
	LEFT_DIR = 0
	RIGHT_DIR = 1
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
	currHead := m.head
	m.mu.Unlock()
	return currHead
}

func (m *MembersList) SetHead(newHead *Node) {
	m.mu.Lock()
	m.head = newHead
	m.mu.Unlock()
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
	for node != nil {
		if(node.status == ALIVE) {
			member := "Machine Id: " + node.ipAddress
			list = append(list, member)
		}
		node = node.left
		if node == m.head {
			break
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
	id := newNode.GetId()
	m.mu.Lock()
	if(m.head == nil) {
		m.head = newNode
		newNode.right = m.head
		newNode.left = m.head
	} else{
		newNode.right = m.head.right
		newNode.left = m.head

		m.head.right.left = newNode
		m.head.right = newNode
	}
	m.nodeMap[id] = newNode
	m.mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
	node := m.nodeMap[id]
	
	if node != nil {
		if len(m.nodeMap) == 1 {
			m.head = nil
		} else {
			// Choose new head if we remove current head
			if node == m.head {
				m.head = m.head.right
			}

			node.left.right = node.right
			node.right.left = node.left
		}
		node.left = nil
		node.right = nil

		delete(m.nodeMap, id)
	}
	m.mu.Unlock()
}

func (m *MembersList) Left(n *Node) *Node {
	m.mu.Lock()
	next := n.left
	m.mu.Unlock()
	return next
}

func (m *MembersList) Right(n *Node) *Node {
	m.mu.Lock()
	next := n.right
	m.mu.Unlock()
	return next
}

func (m *MembersList) GetNeighbor(num int, currNode *Node, direction int) int {
	m.mu.Lock()
	neighbor := currNode
	if direction == RIGHT_DIR {
		for i := -1; i < num; i++ {
			neighbor = neighbor.right
			for neighbor.status != ALIVE {
				neighbor = neighbor.right
			}
			if neighbor == currNode {
				m.mu.Unlock()
				return 0
			}
		}
	} else {
		for i := -1; i < num; i++ {
			neighbor = neighbor.left
			for neighbor.status != ALIVE {
				neighbor = neighbor.left
			}
			if neighbor == currNode {
				m.mu.Unlock()
				return 0
			}
		}
	}
	m.mu.Unlock()
	return neighbor.id
}
