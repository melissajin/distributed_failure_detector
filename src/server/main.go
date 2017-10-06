package main

import (
	"time"
	"net"
	"strconv"
	"os"
	"regexp"
	"strings"
	"bufio"
	"fmt"
//	. "heartbeat"
//	. "membersList"
	"encoding/gob"
	"log"
	"bytes"
	"sync"
)

// Status of machine
const (
	UPDATE = iota	// 0
	JOIN			// 1
)

type Heartbeat struct {
	Id int
	MembershipList MembersList
	Status int
}

func NewHeartbeat(id int, membershipList MembersList, status int) *Heartbeat {
	return &Heartbeat{id, membershipList, status}
}

func (hb *Heartbeat) GetId() int {
	return hb.Id
}

func (hb *Heartbeat) GetMembershipList() MembersList {
	return hb.MembershipList
}

func (hb *Heartbeat) GetStatus() int {
	return hb.Status
}

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
	m.mu.Unlock()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
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
	m.mu.Unlock()
}


// Status of machine
const (
	ALIVE = iota	// 0
	LEAVE			// 1
	FAILED			// 2
)

type Node struct {
	Id int
	HbCounter int
	Timestamp time.Time
	RNeighbor *Node
	RrNeighbor *Node
	LNeighbor *Node
	LlNeighbor *Node
	Status int
}

func NewNode(id int, hbCount int, timestamp time.Time, rNeighbor *Node, rrNeighbor *Node, lNeighbor *Node, llNeighbor *Node, status int) * Node {
	return &Node{ id, hbCount, timestamp, rNeighbor, rrNeighbor, lNeighbor, llNeighbor, status }
}

func (n *Node) GetId() int {
	return n.Id
}

func (n *Node) GetHBCount() int {
	return n.HbCounter
}

func (n *Node) GetTimestamp() time.Time {
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




var memberList MembersList
var leave bool

const (
	connections = 4
	cleanupTime = time.Second * 6		//seconds
	detectionTime = time.Second * 2		//seconds
	heartbeatInterval = time.Second * 1 //seconds
	entryMachineId = 1
)

func main() {
	memberList = NewMembershipList()
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Command: ")
		command, _, _ := reader.ReadLine()
		text := string(command)
		//text, _ := reader.ReadString('\n')
		if(strings.Contains(text, "join")) {
			go Join() //TODO: might need to use thread
		} else if(strings.Contains(text, "leave")) {
			go Leave() //TODO: might need to use thread
		} else if(strings.Contains(text, "list")) {
			if(memberList.Size() == 0){
				fmt.Print("No members\n")
			} else {
				list := memberList.Read()
				fmt.Println(list)
			}
		} else if(strings.Contains(text, "id")) {
			_, id := GetIdentity()
			idStr := strconv.Itoa(id)
			fmt.Print(idStr + "\n")
		} else {
			fmt.Println("Invalid Command. Enter [join/leave/list/id]")
		}
	}
}

func Listen(port int) {
	_, id := GetIdentity()
	addr := getReceiverHost(id, port)

	udpAddr,err := net.ResolveUDPAddr("udp",addr)
	if err != nil {
		log.Fatal("Error getting UDP address:", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)

	for {
		if(leave == true) {
			break
		}

		if(memberList.Size() < 6 && id != entryMachineId) {
			continue
		}

		select {
		case <- time.After(detectionTime):
			currNode := memberList.GetNode(id)
			failedId := getNeighbor(port-8000, currNode)
			failedNode := memberList.GetNode(failedId)

			failedNode.SetStatus(FAILED)
			failedNode.IncrementHBCounter()

			go Cleanup(failedId)

		default:
			// accept and read heartbeat struct from server
			buffer := make([]byte, 1024)
			n, _, _ := conn.ReadFromUDP(buffer)
			if n == 0 {
				fmt.Println("FUCK")
			}
			fmt.Println(n)
			fmt.Println(buffer)
			network := bytes.NewBuffer(buffer)
			fmt.Println(network)
			dec := gob.NewDecoder(network)
			hb := &Heartbeat{}
			err = dec.Decode(hb)
			if err != nil {
				log.Fatal("decode error:", err)
			}
			fmt.Println(hb)
			/*hbStatus := hb.GetStatus()
			receivedMembershipList := hb.GetMembershipList()
			UpdateMembershipLists(receivedMembershipList)

			if(hbStatus == JOIN && id == entryMachineId) {
				// Send hb to new node with current membership list
				entryHB := NewHeartbeat(id, memberList, UPDATE)
				newMachineAddr := getReceiverHost(id, 8000)
				SendOnce(entryHB, newMachineAddr)
			}*/
			conn.Close()
		}
	}
}

func getReceiverHost(machineNum int, portNum int) string {
	var machineStr string
	if(machineNum < 10) {
		machineStr = "0" + strconv.Itoa(machineNum)
	}
	return "fa17-cs425-g46-" + machineStr +".cs.illinois.edu:" + strconv.Itoa(portNum)
}

//Cleanup after clean up period
func Cleanup(id int) {
	time.After(cleanupTime)
	memberList.Remove(id)

	// Kill goroutines for sending and recieving heartbeats
	_, currId := GetIdentity()
	if(currId == id){
		leave = true
	}
}

func UpdateMembershipLists(receivedList MembersList) {
	receivedNode := receivedList.GetHead()
	count := 0
	for count < receivedList.Size() {
		//get node info from their membership list
		receivedStatus := receivedNode.GetStatus()
		receivedHbCount := receivedNode.GetHBCount()
		receivedId := receivedNode.GetId()

		currNode := memberList.GetNode(receivedId)

		if (currNode == nil && receivedStatus == ALIVE) {
			memberList.Insert(receivedNode)
		} else {
			// currStatus := currNode.GetStatus()
			currHBCount := currNode.GetHBCount()

			if(currHBCount < receivedHbCount) {
				if(receivedStatus == ALIVE) {
					r, rr, l, ll := receivedNode.GetNeighbors()
					currNode.SetHBCounter(receivedHbCount)
					currNode.SetNeighbors(r, rr, l, ll)
					currNode.SetStatus(receivedStatus)
				} else if (receivedStatus == LEAVE || receivedStatus == FAILED) {
					currNode.SetHBCounter(receivedHbCount)
					currNode.SetStatus(receivedStatus)
					go Cleanup(receivedId)
				}
			}
		}

		receivedNode = receivedNode.Next()
		count++
	}
}

func GetIdentity() (string, int) {
	host, _ := os.Hostname()
	re, _ := regexp.Compile("-[0-9]+.")
	idStr := re.FindString(host)
	if idStr != "" {
		idStr = idStr[1:len(idStr)-1]
	}
	id, _ := strconv.Atoi(idStr)
	return host, id
}

func Gossip(port int, id int) {
	currNode := memberList.GetNode(id)
	for(memberList.Size() < 6) {}
	receiverId := getNeighbor(port - 8000, currNode)
	receiverAddr := getReceiverHost(receiverId, port)

	for {
		if(leave == true) {
			break
		}

		//send heartbeat after certain duration
		time.After(heartbeatInterval)

		//increment heartbeat counter for node sending hb
		currNode.IncrementHBCounter()

		hb := NewHeartbeat(id, memberList, UPDATE)
		SendOnce(hb, receiverAddr)
	}
}

func SendOnce(hb *Heartbeat, addr string) {
	fmt.Printf("SENDONCE %x %s\n", hb, addr)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}
	enc := gob.NewEncoder(conn)
	e := enc.Encode(hb)
	if e != nil {
		log.Fatal("encode error:", e)
	}
	//conn.Write(network.Bytes())
	conn.Close()
}

func Join() {
	_, id := GetIdentity()

	// Create node and membership list and entry heartbeat
	node := NewNode(id, 0, time.Now(), nil, nil, nil, nil, ALIVE)
	memberList.Insert(node)

	// Get membership list from entry machine
	if(id != entryMachineId) {
		entryHB := NewHeartbeat(id, memberList, JOIN)

		// Send entry heartbeat to entry machine
		entryMachineAddr := getReceiverHost(entryMachineId, 8000)
		SendOnce(entryHB, entryMachineAddr)

		//receive heartbeat from entry machine and update memberList
		udpAddr,err := net.ResolveUDPAddr("udp", entryMachineAddr)
		if err != nil {
			log.Fatal("Error getting UDP address:", err)
		}

		conn, err := net.ListenUDP("udp", udpAddr)
		buffer := make([]byte, 1024)
		n, _, _ := conn.ReadFromUDP(buffer)
		if n == 0 {
			fmt.Println("FUCK2")
		}
		network := bytes.NewBuffer(buffer)
		dec := gob.NewDecoder(network)
		var heartbeat Heartbeat
		err = dec.Decode(&heartbeat)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		conn.Close()

		//merge membership lists
		membershipList := heartbeat.GetMembershipList()
		UpdateMembershipLists(membershipList)
	}

	// start 2 threads for each connection, each listening to different port
	for i := 0; i < connections; i++ {
		leave = false
		go Listen(8000 + i)
		go Gossip(8000 + i, id)
	}
}

func getNeighbor(num int, currNode *Node) int {
	r, rr, l, ll := currNode.GetNeighbors()
	var neighbor *Node
	if(num == 0) {
		neighbor = rr
	} else if(num == 1) {
		neighbor = r
	} else if(num == 2) {
		neighbor = l
	} else if(num == 3) {
		neighbor = ll
	}

	return neighbor.GetId()
}

func Leave() {
	//remove self from membership list
	_, id := GetIdentity()
	leaveNode := memberList.GetNode(id)

	leaveNode.SetStatus(LEAVE)
	go Cleanup(id)
}

func printMemberList() {
	currNode := memberList.GetHead()
	if(currNode == nil){
		fmt.Println("No members to print")
	} else {
		fmt.Printf("Node id: %d\n", currNode.GetId())
	}
}