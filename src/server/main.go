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
	"sync"
)

var memberList MembersList
var leave chan bool
var entryMachineIds = []int{1,2,3,4,5}

const (
	connections = 4
	cleanupTime = time.Second * 6
	detectionTime = time.Second * 2
	heartbeatInterval = time.Millisecond * 100
)

func main() {
	memberList = NewMembershipList()
	_, id := GetIdentity()

	// Create logfile
	logfileName := "machine." + strconv.Itoa(id) + ".log"
	f, err := os.OpenFile(logfileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error opening log file: ", err)
	}
	defer f.Close()
	log.SetOutput(f)

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

func Listen(port int, wg *sync.WaitGroup) {
	defer wg.Done()

	_, id := GetIdentity()
	addr := getReceiverHost(id, port)

	udpAddr,err := net.ResolveUDPAddr("udp",addr)
	if err != nil {
		log.Fatal("Error getting UDP address:", err)
	}

	if err != nil {
		log.Fatal("Error listening:", err)
	}
	ListenLoop:
		for {
			if memberList.Size() < 2 && !Contains(entryMachineIds, id) {
				continue
			}

			select {
			case <- leave:
				break ListenLoop
			default:
				buffer := make([]byte, 1024)
				conn, err := net.ListenUDP("udp", udpAddr)
				//conn.SetReadDeadline(time.Now().Add(detectionTime))
				if err != nil {
					fmt.Println("ERROR: ", err, conn)
				}
				_ , _, err = conn.ReadFrom(buffer)
				conn.Close()
				//if err != nil {
				//	fmt.Println("ERROR:", err)
				//	currNode := memberList.GetNode(id)
				//	fmt.Println(currNode, memberList, id)
				//	failedId := getNeighbor(port-8000, currNode)
				//
				//	if failedId == 0 {
				//		continue
				//	}
				//	failedNode := memberList.GetNode(failedId)
				//	failedNode.SetStatus(FAILED)
				//	failedNode.IncrementHBCounter()
				//	log.Printf("Machine %d failed at port %d", failedId, port)
				//	go Cleanup(failedId)
				//	continue
				//}

				buffer = bytes.Trim(buffer, "\x00")
				if(len(buffer) == 0){
					continue
				}
				hb := &pb.Heartbeat{}
				err = proto.Unmarshal(buffer, hb)
				if err != nil {
					log.Fatal("Unmarshal error:", err)
				}

				receivedMembershipList := hb.GetMachine()
				receivedMachineId := int(hb.GetId())
				log.Println("Update Membership List")
				UpdateMembershipLists(receivedMembershipList)
				if(len(receivedMembershipList) == 1 && Contains(entryMachineIds, id)) {
					// Send hb to new node with current membership list
					entryHB := ConstructPBHeartbeat()
					newMachineAddr := getReceiverHost(receivedMachineId, 8000)
					log.Println(id, " entry send to ", newMachineAddr)
					SendOnce(entryHB, newMachineAddr)
				}
			}
		}
}

func Contains(arr []int, num int) bool {
	for i := 0; i < len(entryMachineIds); i++ {
		if num == entryMachineIds[i] {
			return true
		}
	}
	return false
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
	time.Sleep(cleanupTime)
	_, ownId := GetIdentity()

	if(id == ownId) {
		// Reset membership list
		memberList = NewMembershipList()
		log.Println("Reset membership list")
	} else {
		memberList.Remove(id)
	}
}

func UpdateMembershipLists(receivedList []*pb.Machine) {
	recievedMemList := NewMembershipList()

	// Construct Membership List from received heartbeat
	for i := 0; i < len(receivedList); i++ {
		machine := receivedList[i]
		receivedId := machine.GetId()
		receivedStatus := int(machine.GetStatus())
		receivedHbCount := int(machine.GetHbCounter())
		newNode := NewNode(int(receivedId.Id), receivedHbCount, receivedId.Timestamp, receivedStatus)
		recievedMemList.Insert(newNode)
	}

	if memberList.Size() == 1 {
		MergeLists(recievedMemList, memberList)
	} else {
		MergeLists(memberList, recievedMemList)
	}
}

// Merges the list B into the list A
func MergeLists(A MembersList, B MembersList) MembersList {
	currB := B.GetHead()

	for currB != nil {
		currA := A.GetNode(currB.GetId())
		statusB := currB.GetStatus()
		idB := currB.GetId()
		hbCountB := currB.GetHBCount()
		timestampB := currB.GetTimestamp()

		if currA == nil {
			if statusB == ALIVE {
				log.Printf("Machine %d joined", idB)
				newNode := NewNode(idB, hbCountB, timestampB, statusB)
				A.Insert(newNode)
			}
		} else {
			hbCountA := currA.GetHBCount()
			statusA := currA.GetStatus()
			timestampA := currA.GetTimestamp()

			if hbCountA < hbCountB {
				// Log falsely detected failure
				if statusA == FAILED && statusB == ALIVE && timestampA == timestampB {
					log.Println("Falsely detected failure at machine ", idB)
				} else {
					if statusB == LEAVE || statusB == FAILED {
						go Cleanup(idB)
					}
					currA.SetHBCounter(hbCountB)
					currA.SetStatus(statusB)
				}
			}
		}
		currB = currB.Next()
		if currB == B.GetHead() {
			break
		}
	}
	return A
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

func Gossip(port int, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	currNode := memberList.GetNode(id)
	for(memberList.Size() < 2) {}

	GossipLoop:
		for {
			select {
			case <- leave:
				break GossipLoop
			default:
				//send heartbeat after certain duration
				time.Sleep(heartbeatInterval)

				receiverId := getNeighbor(port - 8000, currNode)
				if receiverId == 0 {
					continue
				}
				receiverAddr := getReceiverHost(receiverId, port)

				//increment heartbeat counter for node sending hb
				currNode.IncrementHBCounter()

				hb := ConstructPBHeartbeat()
				log.Println(id, " gossip to ", receiverAddr)
				SendOnce(hb, receiverAddr)
			}
		}
}

func SendOnce(hb *pb.Heartbeat, addr string) {
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
	head := memberList.GetHead()
	node := head
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
		if node == head {
			break
		}
	}

	return hb
}

func Join() {
	leave = make(chan bool)
	_, id := GetIdentity()
	// Create node and membership list and entry heartbeat
	node := NewNode(id, 0, time.Now().String(), ALIVE)
	memberList.Insert(node)

	var wg sync.WaitGroup
	wg.Add(5)
	// Get membership list from one entry machine
	for i := 0; i < len(entryMachineIds); i++ {
		go GetCurrentMembers(entryMachineIds[i], &wg)
	}
	wg.Wait()

	wg.Add(3)
	// start 2 threads for each connection, each listening to different port
	for i := 0; i < connections; i++ {
		go Listen(8000 + i, &wg)
		go Gossip(8000 + i, id, &wg)
	}
	wg.Wait()

	log.Println("Last cleanup")
	Cleanup(id)
}

func GetCurrentMembers(entryId int, wg *sync.WaitGroup) {
	defer wg.Done()

	_, id := GetIdentity()
	if id == entryId {
		return
	}

	entryHB := ConstructPBHeartbeat()

	// Send entry heartbeat to entry machine
	entryMachineAddr := getReceiverHost(entryId, 8000)
	log.Println(entryId, " ask to join ", entryMachineAddr)
	SendOnce(entryHB, entryMachineAddr)

	//receive heartbeat from entry machine and update memberList
	receiverMachineAddr := getReceiverHost(id, 8000)
	udpAddr,err := net.ResolveUDPAddr("udp", receiverMachineAddr)
	if err != nil {
		log.Fatal("Error getting UDP address:", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("Error listening to addr: ", err)
		return
	}
	//defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(detectionTime))
	buffer := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buffer)
	conn.Close()
	if err != nil {
		return
	}
	buffer = bytes.Trim(buffer, "\x00")
	hb := &pb.Heartbeat{}
	err = proto.Unmarshal(buffer, hb)
	if err != nil {
		log.Fatal("Unmarshal2 error:", err)
	}

	//merge membership lists
	UpdateMembershipLists(hb.Machine)
}

func getNeighbor(num int, currNode *Node) int {
	r, rr, l, ll := currNode.GetNeighbors()
	_, id := GetIdentity()
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

	if neighbor != nil && neighbor.GetId() != id {
		return neighbor.GetId()
	} else {
		return 0
	}
}

func Leave() {
	_, id := GetIdentity()

	//remove self from membership list
	leaveNode := memberList.GetNode(id)
	leaveNode.SetStatus(LEAVE)

	log.Printf("Machine %d left", id)
	// Kill goroutines for sending and receiving heartbeats
	close(leave)
}