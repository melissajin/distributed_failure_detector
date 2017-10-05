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
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	addr := host + ":" + strconv.Itoa(port)

	// Don't start gossip until > 5 machines in the system or if machine is entry machine
	for memberList.Size() < 6 {
		if(id == entryMachineId) {
			break
		}
	}

	ln, err := net.Listen("udp", addr)
	if err != nil {
		log.Fatal("Error when listening to port:", err)
	}
	for {
		if(leave == true) {
			break
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
			conn, err := ln.Accept()
			if err != nil {
				log.Fatal("Error when accepting connection:", err)
			}

			dec := gob.NewDecoder(conn)
			var hb Heartbeat
			err = dec.Decode(&hb)
			if err != nil {
				log.Fatal("decode error:", err)
			}

			hbStatus := hb.GetStatus()
			receivedMembershipList := hb.GetMembershipList()
			receivedId := hb.GetId()
			receivedIdStr := strconv.Itoa(receivedId)
			UpdateMembershipLists(receivedMembershipList)

			if(hbStatus == JOIN && id == entryMachineId) {
				// Send hb to new node with current membership list
				entryHB := NewHeartbeat(id, memberList, UPDATE)
				newMachineAddr := strings.Replace(host, idStr, receivedIdStr, -1) + ":" + strconv.Itoa(8000)
				SendOnce(entryHB, newMachineAddr)
			} 
			conn.Close()
		}
	}
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

func Gossip(port int, host string, id int) {
	addr := host + ":" + strconv.Itoa(port)
	for {
		if(leave == true) {
			break
		}

		// Don't start gossip until > 5 machines in the system
		if(memberList.Size() < 6){
			continue
		}

		//send heartbeat after certain duration
		time.After(heartbeatInterval)

		//increment heartbeat counter for node sending hb
		currNode := memberList.GetNode(id)
		currNode.IncrementHBCounter()

		hb := NewHeartbeat(id, memberList, UPDATE)
		SendOnce(hb, addr)
	}
}

func SendOnce(hb *Heartbeat, addr string) {
	fmt.Printf("SENDONCE %x %s\n", hb, addr)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}
	enc := gob.NewEncoder(conn)
	//conn.Write([]byte(hb))
	e := enc.Encode(*hb)
	if e != nil {
		log.Fatal("encode error:", err)
	}
	conn.Close()
}

func Join() {
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	entryMachineIdStr := strconv.Itoa(entryMachineId)

	// Create node and membership list and entry heartbeat
	node := NewNode(id, 0, time.Now(), nil, nil, nil, nil, ALIVE)
	memberList.Insert(node)

	// Get membership list from entry machine
	if(id != entryMachineId) {
		entryHB := NewHeartbeat(id, memberList, JOIN)

		// Send entry heartbeat to entry machine
		entryMachineAddr := strings.Replace(host, idStr, entryMachineIdStr, -1) + ":" + strconv.Itoa(8000)
		SendOnce(entryHB, entryMachineAddr)

		//receive heartbeat from entry machine and update memberList
		ln, err := net.Listen("udp", entryMachineAddr)
		conn, err := ln.Accept()
		dec := gob.NewDecoder(conn)
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
		receiverId := getNeighbor(i, node)
		recieverIdStr := strconv.Itoa(receiverId)
		receiverHost := strings.Replace(host, idStr, recieverIdStr, -1)
		leave = false
		go Listen(8000 + i)
		go Gossip(8000 + i, receiverHost, id)
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