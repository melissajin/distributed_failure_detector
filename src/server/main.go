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
	. "membersList"
	"log"
	"github.com/golang/protobuf/proto"
	pb "heartbeat/heartbeat"
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

	if err != nil {
		log.Fatal("Error listening:", err)
	}

	for {
		if(leave == true) {
			break
		}
		fmt.Println("LISTEN 1")
		if(memberList.Size() < 2 && id != entryMachineId) {
			continue
		}
		fmt.Println("LISTEN 2")
		select {
		case <- time.After(detectionTime):
			fmt.Println("LISTEN 3")
			currNode := memberList.GetNode(id)
			failedId := getNeighbor(port-8000, currNode)
			if failedId == 0 {
				continue
			}
			fmt.Println("LISTEN 4")
			failedNode := memberList.GetNode(failedId)
			failedNode.SetStatus(FAILED)
			failedNode.IncrementHBCounter()

			go Cleanup(failedId)

		default:
			fmt.Println("LISTEN 5")
			buffer := make([]byte, 1024)
			conn, err := net.ListenUDP("udp", udpAddr)
			conn.ReadFrom(buffer)
			conn.Close()

			buffer = bytes.Trim(buffer, "\x00")
			if(len(buffer) == 0){
				continue
			}
			fmt.Println("LISTEN 6")
			hb := &pb.Heartbeat{}
			err = proto.Unmarshal(buffer, hb)
			if err != nil {
				log.Fatal("Unmarshal error:", err)
			}

			receivedMembershipList := hb.GetMachine()
			receivedMachindId := int(hb.GetId())
			UpdateMembershipLists(receivedMembershipList)
			fmt.Println(receivedMembershipList)
			fmt.Println("LISTEN 7")
			if(len(receivedMembershipList) == 1 && id == entryMachineId) {
				// Send hb to new node with current membership list
				entryHB := ConstructPBHeartbeat()
				newMachineAddr := getReceiverHost(receivedMachindId, 8000)
				SendOnce(entryHB, newMachineAddr)
			}
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

func UpdateMembershipLists(receivedList []*pb.Machine) {
	for i := 0; i < len(receivedList); i++ {
		machine := receivedList[i]
		//get node info from their membership list
		receivedId := machine.GetId()
		nodeId := int(receivedId.Id)
		receivedStatus := int(machine.GetStatus())
		receivedHbCount := int(machine.GetHbCounter())

		currNode := memberList.GetNode(nodeId)

		if currNode == nil && receivedStatus == ALIVE {
			newNode := NewNode(int(receivedId.Id), receivedHbCount, receivedId.Timestamp, nil, nil, nil, nil, receivedStatus)
			memberList.Insert(newNode)
		} else {
			// currStatus := currNode.GetStatus()
			currHBCount := currNode.GetHBCount()

			if currHBCount < receivedHbCount {
				if receivedStatus == LEAVE || receivedStatus == FAILED {
					go Cleanup(nodeId)
				}
				currNode.SetHBCounter(receivedHbCount)
				currNode.SetStatus(receivedStatus)
			}
		}
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
	for(memberList.Size() < 2) {}

	for {
		if(leave == true) {
			break
		}

		//send heartbeat after certain duration
		time.Sleep(heartbeatInterval)

		fmt.Printf("GOSSIP 1 %d", port)
		receiverId := getNeighbor(port - 8000, currNode)
		if receiverId == 0 {
			continue
		}
		fmt.Println("GOSSIP 2")
		receiverAddr := getReceiverHost(receiverId, port)

		//increment heartbeat counter for node sending hb
		currNode.IncrementHBCounter()

		hb := ConstructPBHeartbeat()
		SendOnce(hb, receiverAddr)
	}
}

func SendOnce(hb *pb.Heartbeat, addr string) {
	fmt.Printf("SENDONCE %x %s\n", hb, addr)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}
	out, err := proto.Marshal(hb)
	if err != nil {
		log.Fatal("Marshal error:", err)
	}
	conn.Write(out)
	conn.Close()
}

func ConstructPBHeartbeat() *pb.Heartbeat{
	_, id := GetIdentity()
	hb := &pb.Heartbeat{}
	hb.Id = int32(id)
	node := memberList.GetHead()
	for node != nil {
		machine := &pb.Machine{}
		machineId := &pb.Machine_Id{}
		machineId.Id = int32(node.GetId())
		machineId.Timestamp = node.GetTimestamp()
		machine.HbCounter = int32(node.GetHBCount())
		machine.Status = int32(node.GetStatus())
		machine.Id = machineId
		hb.Machine = append(hb.Machine, machine)

		node = node.Next()
	}

	return hb
}

func Join() {
	_, id := GetIdentity()

	// Create node and membership list and entry heartbeat
	node := NewNode(id, 0, time.Now().String(), nil, nil, nil, nil, ALIVE)
	memberList.Insert(node)

	// Get membership list from entry machine
	if(id != entryMachineId) {
		entryHB := ConstructPBHeartbeat()

		// Send entry heartbeat to entry machine
		entryMachineAddr := getReceiverHost(entryMachineId, 8000)
		SendOnce(entryHB, entryMachineAddr)

		//receive heartbeat from entry machine and update memberList
		receiverMachineAddr := getReceiverHost(id, 8000)
		udpAddr,err := net.ResolveUDPAddr("udp", receiverMachineAddr)
		if err != nil {
			log.Fatal("Error getting UDP address:", err)
		}

		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			log.Fatal("Error listening to addr: ", err)
		}
		buffer := make([]byte, 1024)
		conn.ReadFromUDP(buffer)
		buffer = bytes.Trim(buffer, "\x00")
		hb := &pb.Heartbeat{}
		err = proto.Unmarshal(buffer, hb)
		if err != nil {
			log.Fatal("Unmarshal2 error:", err)
		}
		conn.Close()

		//merge membership lists
		UpdateMembershipLists(hb.Machine)
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

	if neighbor != nil {
		return neighbor.GetId()
	} else {
		return 0
	}
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