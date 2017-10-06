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
	. "heartbeat"
	. "membersList"
	"encoding/gob"
	"log"
	"bytes"
)

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
			fmt.Println(hb)
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