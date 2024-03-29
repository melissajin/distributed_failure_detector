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
	"math"
	"math/rand"
)

var memberList MembersList
var leave chan bool
var entryMachineIds = []int{1,2,3,4,5}

const (
	CONNECTIONS = 4
	CLEANUP_TIME = time.Second * 2
	DETECTION_TIME = time.Second * 2
	STARTUP_TIME = time.Second * 1
	HB_INTERVAL = time.Millisecond * 400
	MESSAGE_LOSS_RATE = 0
)

type Counter struct {
    mu  sync.Mutex
    x   int
}

/**
 * Atomic add for type Counter
 */
func (c *Counter) Add(x int) {
    c.mu.Lock()
    c.x += x
    c.mu.Unlock()
}

func(c *Counter) Reset() {
	c.mu.Lock()
    c.x = 0
    c.mu.Unlock()
}

/* Global variables to keep track of total messages sent and recieved at a machine */
var messagesRecieved Counter
var messagesSent Counter

/* used to simulate message loss */
func simMessageLoss(percentage int) bool {
	if percentage <= 0 {
		return true
	}

	failOn := 100 / percentage

	s1 := rand.NewSource(time.Now().UnixNano())
    r1 := rand.New(s1)
    randNum := r1.Intn(failOn)

	if(randNum == failOn-1) {
		return false
	}

	return true
}


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
			go Join()
		} else if(strings.Contains(text, "leave")) {
			go Leave()
		} else if(strings.Contains(text, "list")) {
			_, id = GetIdentity()
			node := memberList.GetNode(id)
			if node == nil || node.GetStatus() != ALIVE {
				fmt.Print("Not joined\n")
			} else {
				list := memberList.Read()
				fmt.Println(list)
			}
		} else if(strings.Contains(text, "id")) {
			ip, _ := os.Hostname()
			fmt.Print(ip + "\n")
		} else {
			fmt.Println("Invalid Command. Enter [join/leave/list/id]")
		}
	}
}

func Listen(port int, wg *sync.WaitGroup) {
	defer wg.Done()

	_, id := GetIdentity()
	addr := getAddress(id, port)

	udpAddr,err := net.ResolveUDPAddr("udp",addr)
	if err != nil {
		log.Fatal("Error getting UDP address:", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("Error listening: ", err)
	}

	ListenLoop:
		for {
			select {
			case <- leave:
				break ListenLoop
			default:
				buffer := make([]byte, 1024)
				currNode := memberList.GetNode(id)
				if currNode == nil || currNode.GetStatus() == FAILED{
					break ListenLoop
				}
				neighbor := memberList.GetNeighbor(port-8000, currNode, RIGHT_DIR)
				// TODO: only set deadline after machine joins
				if(neighbor != 0) {
					neighNode := memberList.GetNode(neighbor)
					if neighNode == nil || neighNode.GetStatus() != ALIVE {
						continue
					}
					
					conn.SetReadDeadline(time.Now().Add(DETECTION_TIME))
					_ , _, err = conn.ReadFrom(buffer)
					// Timeout error, machine failed
					if err != nil {
						if err, ok := err.(net.Error); ok && err.Timeout() {
							currNode := memberList.GetNode(id)
							if currNode == nil {
								continue
							}
							failedId := memberList.GetNeighbor(port-8000, currNode, RIGHT_DIR)
							if failedId == 0 {
								continue
							}
							log.Printf("Machine %d failed at port %d", failedId, port)
							failedNode := memberList.GetNode(failedId)
							failedNode.SetStatus(FAILED)
							failedNode.IncrementHBCounter()
							hb := ConstructPBHeartbeat()
							SendOnce(hb, getAddress(failedId, port))
							go Cleanup(failedId)
							conn.SetReadDeadline(time.Now().Add(DETECTION_TIME))
							continue
						} else {
							continue
						}
					}
				}

				buffer = bytes.Trim(buffer, "\x00")
				if(len(buffer) == 0){
					continue
				}

				messagesRecieved.Add(1)
				log.Println("LISTEN: received hb from ", neighbor, " on ", port, "Total recieved: ", messagesRecieved.x)
				hb := &pb.Heartbeat{}
				err = proto.Unmarshal(buffer, hb)
				if err != nil {
					log.Fatal("Unmarshal error:", err)
				}

				receivedMembershipList := hb.GetMachine()
				UpdateMembershipLists(receivedMembershipList)
			}
		}
	conn.Close()

}

func Contains(arr []int, num int) bool {
	for i := 0; i < len(entryMachineIds); i++ {
		if num == entryMachineIds[i] {
			return true
		}
	}
	return false
}

func getAddress(machineNum int, portNum int) string {
	machineStr := strconv.Itoa(machineNum)
	if(machineNum < 10) {
		machineStr = "0" + machineStr
	}
	return "fa17-cs425-g46-" + machineStr +".cs.illinois.edu:" + strconv.Itoa(portNum)
}

func getHost(machineNum int) string {
	machineStr := strconv.Itoa(machineNum)
	if(machineNum < 10) {
		machineStr = "0" + machineStr
	}
	return "fa17-cs425-g46-" + machineStr +".cs.illinois.edu"
}

//Cleanup after clean up period
func Cleanup(id int) {
	time.Sleep(CLEANUP_TIME)
	_, ownId := GetIdentity()

	if(id == ownId) {
		// Reset membership list
		memberList = NewMembershipList()
		log.Println("Reset membership list")
	} else {
		memberList.Remove(id)
		log.Printf("Remove %d from list", id)
	}
}

func UpdateMembershipLists(receivedList []*pb.Machine) {

	recievedMemList := NewMembershipList()

	// Construct Membership List from received heartbeat
	for i := 0; i < len(receivedList); i++ {
		machine := receivedList[i]
		receivedId := machine.GetId()
		receivedIp := getHost(int(receivedId.Id))
		receivedStatus := int(machine.GetStatus())
		receivedHbCount := int(machine.GetHbCounter())
		newNode := NewNode(int(receivedId.Id), receivedIp, receivedHbCount, receivedId.Timestamp, receivedStatus)
		recievedMemList.Insert(newNode)
	}

	if memberList.Size() == 1 && recievedMemList.Size() != 1 {
		id := memberList.GetHead().GetId()
		memberList = recievedMemList
		log.Printf("Machine %d joined", id)
	} else {
		memberList = MergeLists(memberList, recievedMemList)
	}
}

// Merges the list B into the list A
func MergeLists(A MembersList, B MembersList) MembersList {
	currB := B.GetHead()
	for currB != nil {
		currA := A.GetNode(currB.GetId())
		statusB := currB.GetStatus()
		idB := currB.GetId()
		ipAddressB := getHost(idB)
		hbCountB := currB.GetHBCount()
		timestampB := currB.GetTimestamp()
		if currA == nil {
			if statusB == ALIVE {
				log.Printf("Machine %d joined", idB)
				newNode := NewNode(idB, ipAddressB ,hbCountB, timestampB, statusB)
				A.Insert(newNode)
			}
		} else {
			hbCountA := currA.GetHBCount()
			statusA := currA.GetStatus()
			timestampA := currA.GetTimestamp()

			if timestampB == timestampA {
				if statusA == ALIVE {
					if statusB == LEAVE {
						log.Printf("Machine %d left", idB)
						currA.SetStatus(statusB)
						go Cleanup(idB)
					} else if statusB == FAILED {
						log.Printf("Detected failure at machine %d", idB)
						currA.SetStatus(statusB)
						go Cleanup(idB)
					} else if statusB == ALIVE {
						currA.SetHBCounter(int(math.Max(float64(hbCountA), float64(hbCountB))))
					}
				} else if statusB == ALIVE {
						log.Println("Falsely detected failure at machine ", idB)
					}
			}
		}
		currB = B.Left(currB)
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

	GossipLoop:
		for {
			for(memberList.Size() < 2) {}
			select {
			case <- leave:
				break GossipLoop
			default:
				//send heartbeat after certain duration
				time.Sleep(HB_INTERVAL)

				if currNode.GetStatus() == FAILED {
					break GossipLoop
				}
				receiverId := memberList.GetNeighbor(port - 8000, currNode, LEFT_DIR)
				if receiverId == 0 {
					continue
				}
				receiverNode := memberList.GetNode(receiverId)
				if receiverNode != nil && receiverNode.GetStatus() == ALIVE {
					receiverAddr := getAddress(receiverId, port)

					// Increment heartbeat counter for node sending hb
					currNode.IncrementHBCounter()

					hb := ConstructPBHeartbeat()
					messagesSent.Add(1)
					log.Println("GOSSIP: ", id, "send to", receiverId, "on", port, "Total Sent:", messagesSent.x)
					
					if simMessageLoss(MESSAGE_LOSS_RATE) == true {
						SendOnce(hb, receiverAddr)
					} else {
						log.Println("Dropped packet")
					}
				}
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

		node = memberList.Left(node)
		if node == head {
			break
		}
	}

	return hb
}

func Join() {
	leave = make(chan bool)
	host , id := GetIdentity()

	// Create node and membership list and entry heartbeat
	node := NewNode(id, host,0, time.Now().String(), ALIVE)
	memberList.Insert(node)

	// Get membership list from one entry machine
	var wg1 sync.WaitGroup
	wg1.Add(5)
	for i := 0; i < len(entryMachineIds); i++ {
		go GetCurrentMembers(entryMachineIds[i], &wg1)
	}
	wg1.Wait()

	// start 4 threads to listen and 4 threads to gossip
	var wg2 sync.WaitGroup
	wg2.Add(8)
	for i := 0; i < CONNECTIONS; i++ {
		go Listen(8000 + i, &wg2)
		go Gossip(8000 + i, id, &wg2)
	}

	// Setup ports to listen for new machines if it is entry machine
	if(Contains(entryMachineIds, id)) {
		wg2.Add(1)
		go SetupEntryPort(&wg2)
	}
	wg2.Wait()
	Cleanup(id)
}

func SetupEntryPort(wg *sync.WaitGroup) {
	defer wg.Done()
	_, id := GetIdentity()
	addr := getAddress(id, 8005)

	udpAddr,err := net.ResolveUDPAddr("udp",addr)
	if err != nil {
		log.Fatal("Error getting UDP address:", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)

	EntryLoop:
		for {
			select {
			case <- leave:
				break EntryLoop

			default:
				currNode := memberList.GetNode(id)
				if currNode == nil || currNode.GetStatus() == FAILED {
					break EntryLoop
				}

				buffer := make([]byte, 1024)
				conn.SetReadDeadline(time.Now().Add(DETECTION_TIME))
				_ , _, err = conn.ReadFrom(buffer)
				if err != nil {
					continue
				}

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
				UpdateMembershipLists(receivedMembershipList)
				entryHB := ConstructPBHeartbeat()
				receivedMachineId := int(hb.GetId())
				newMachineAddr := getAddress(receivedMachineId, 8000+id)
				SendOnce(entryHB, newMachineAddr)
			}
		}
	conn.Close()

}

func GetCurrentMembers(entryId int, wg *sync.WaitGroup) {
	defer wg.Done()

	_, id := GetIdentity()
	if id == entryId {
		return
	}

	entryHB := ConstructPBHeartbeat()

	// Send entry heartbeat to entry machine
	entryMachineAddr := getAddress(entryId, 8005)
	log.Println(entryId, " ask to join ", entryMachineAddr)
	SendOnce(entryHB, entryMachineAddr)

	//receive heartbeat from entry machine and update memberList
	receiverMachineAddr := getAddress(id, 8000+entryId)
	udpAddr,err := net.ResolveUDPAddr("udp", receiverMachineAddr)
	if err != nil {
		log.Fatal("Error getting UDP address:", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("Error listening to addr: ", receiverMachineAddr, err)
		conn.Close()
		return
	}

	conn.SetReadDeadline(time.Now().Add(STARTUP_TIME))
	buffer := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buffer)
	conn.Close()
	if err != nil {
		log.Println("Error reading from UPD buffer", receiverMachineAddr, err)
		return
	}
	buffer = bytes.Trim(buffer, "\x00")
	hb := &pb.Heartbeat{}
	err = proto.Unmarshal(buffer, hb)
	if err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	UpdateMembershipLists(hb.Machine)
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