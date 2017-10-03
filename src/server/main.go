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
	. "../node"
	"encoding/gob"
	"log"
)

var memberList MembersList
var hbCounter int
var timestamp time.Time
var leave bool

const (
	connections = 4
	cleanupTime = time.Second * 6		//seconds
	detectionTime = time.Second * 2		//seconds
	heartbeatInterval = time.Second * 1 //seconds
	entryMachineId = 1
)

func main() {
	IdMap = make(map[int]*Node)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Command: ")
		command, _, _ := reader.ReadLine()
		text := string(command)
		//text, _ := reader.ReadString('\n')
		if(strings.EqualFold(text, "join")) {
			go Join() //TODO: might need to use thread
		} else if(strings.EqualFold(text, "leave")) {
			go Leave() //TODO: might need to use thread
		} else if(strings.EqualFold(text, "list")) {
			if(memberList.Length() == 0) {
				fmt.Print("No members\n")
				continue
			}
			fmt.Print(memberList.Read())
		} else if(strings.EqualFold(text, "id")) {
			_, id := GetIdentity()
			idStr := strconv.Itoa(id)
			fmt.Print(idStr + "\n")
		} else {
			fmt.Println("Invalid Command")
		}
	}
}

func Listen(port int, receiverHost string, receiverId string) {
	addr := ":" + strconv.Itoa(port)
	ln, err := net.Listen("udp", addr)
	if err != nil {
		log.Fatal("Error when listening to port:", err)
	}
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	for {
		if(leave == true) {
			break
		}
		select {
		case <- time.After(detectionTime):
			//remove id from list
			removeId, _ := strconv.Atoi(receiverId)

			//heartbeat failure
			var hb Heartbeat
			hb.SetInfo(id, host, memberList, time.Now(), FAILED)
			for i := 0; i < connections; i++ {
				receiverId := strconv.Itoa((id + i + 1) % len(memberList.Read()))
				receiverAddr := strings.Replace(host, idStr, receiverId, -1)
				SendOnce(hb, receiverAddr)
			}

			go Cleanup(removeId)

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

			//merge membership lists
			_, _, membershipList, _, _ := hb.GetInfo()
			UpdateMembershipLists(membershipList)

			conn.Close()
		}
	}
}


//Cleanup after clean up period
func Cleanup(id int) {
	time.After(cleanupTime)
	memberList.Remove(id)
}

func UpdateMembershipLists(membershipList MembersList) {
	receivedNode := membershipList.Head()
	count := 0
	for count < membershipList.Length() {
		//get node info from their membership list
		receivedId, receivedHbCount, receievedTs, receievedRNeighbor, receievedRrNeighbor, receievedLNeighbor, receievedLlNeighbor, receievedStatus := receivedNode.GetInfo()
		node := IdMap[receivedId]
		//get node info from own membership list
		_, hbCount, _, _, _, _, _, status := node.GetInfo()
		if(status == ALIVE) {
			if(hbCount < receivedHbCount) {
				IdMap[receivedId].SetInfo(receivedHbCount, receievedTs, receievedRNeighbor, receievedRrNeighbor, receievedLNeighbor, receievedLlNeighbor, receievedStatus)
			}
		} else if (status == LEAVE) {
			continue
		} else if (node == nil) {
			IdMap[receivedId] = receivedNode
		}
		receivedNode = receievedRNeighbor
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

func Send(port int, host string, id int) {
	addr := host + ":" + strconv.Itoa(port)
	for {
		if(leave == true) {
			break
		}
		//send heartbeat after certain duration
		time.After(heartbeatInterval)

		//increment heartbeat counter of all nodes before sending
			UpdateHeartbeatCounter()

		var hb Heartbeat
		hb.SetInfo(id, host, memberList, timestamp, ALIVE)
		SendOnce(hb, addr)
	}
}

func UpdateHeartbeatCounter() {
	for _, value := range IdMap {
		_, hbCount, timestamp, rNeighbor, rrNeighbor, lNeighbor, llNeighbor, status := value.GetInfo()
		if(status == ALIVE) {
			value.SetHbCounter(hbCount + 1)
		} else if (status == JOIN) {
			value.SetInfo(hbCount + 1, timestamp, rNeighbor, rrNeighbor, lNeighbor, llNeighbor, ALIVE)
		}
	}
}

func SendOnce(hb Heartbeat, addr string) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}
	enc := gob.NewEncoder(conn)
	//conn.Write([]byte(hb))
	e := enc.Encode(hb)
	if e != nil {
		log.Fatal("encode error:", err)
	}
	conn.Close()
}

func Join() {
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	entryMachineIdStr := strconv.Itoa(entryMachineId)
	timestamp = time.Now()

	//add self to membership list and contact fixed entry machine and get membership list
	memberList.Insert(id, timestamp, hbCounter)

	var hb Heartbeat
	hb.SetInfo(id, host, memberList, timestamp, JOIN)
	entryMachineAddr := strings.Replace(host, idStr, entryMachineIdStr, -1)
	SendOnce(hb, entryMachineAddr)

	//receive heartbeat from entry machine and update memberList
	ln, err := net.Listen("udp", entryMachineIdStr)
	conn, err := ln.Accept()
	dec := gob.NewDecoder(conn)
	var heartbeat Heartbeat
	err = dec.Decode(&heartbeat)
	if err != nil {
		log.Fatal("decode error:", err)
	}

	//merge membership lists
	_, _, membershipList, _, _ := heartbeat.GetInfo()
	UpdateMembershipLists(membershipList)

	conn.Close()

	// start 2 threads for each connection, each listening to different port
	for i := 0; i < connections; i++ {
		receiverId := strconv.Itoa((id + i + 1) % len(memberList.Read()))
		receiverHost := strings.Replace(host, idStr, receiverId, -1)
		leave = false
		go Listen(8000 + i, receiverHost, receiverId)
		go Send(8000 + i, receiverHost, id)
	}
}

func Leave() {
	//remove self from membership list
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	hbCounter = IdMap[id].GetHbCounter()
	memberList.Remove(id)

	// Send heartbeat to leave
	var hb Heartbeat
	hb.SetInfo(id, host, memberList, time.Now(), LEAVE)
	for i := 0; i < connections; i++ {
		receiverId := strconv.Itoa((id + i + 1) % len(memberList.Read()))
		receiverAddr := strings.Replace(host, idStr, receiverId, -1)
		SendOnce(hb, receiverAddr)
	}

	// Kill goroutines for sending and recieving heartbeats
	leave = true
}