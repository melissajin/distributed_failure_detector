package membersList

// Status of machine
const (
	ALIVE = iota	// 0
	LEAVE			// 1
	FAILED			// 2
)

type Node struct {
	id int
	ipAddress string
	hbCounter int
	timestamp string
	right *Node
	left *Node
	status int
}

func NewNode(id int, ipAddress string, hbCount int, timestamp string, status int) * Node {
	return &Node{ id, ipAddress, hbCount, timestamp, nil, nil, status }
}

func (n *Node) GetId() int {
	return n.id
}

func (n *Node) GetIPAddress() string {
	return n.ipAddress
}

func (n *Node) GetHBCount() int {
	return n.hbCounter
}

func (n *Node) GetTimestamp() string {
	return n.timestamp
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

func (n *Node) SetNeighbors(right *Node, left *Node) {
	n.right = right
	n.left = left
}

func (n *Node) SetStatus(status int) {
	n.status = status
}
