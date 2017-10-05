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
			list := memberList.Read()
			if(len(list) == 0) {
				fmt.Print("No members\n")
				continue
			} else {
				fmt.Print(list)
			}
		} else if(strings.Contains(text, "id")) {
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
		case <- time.After(detectionTime): // TODO: need to reset detectionTime when recieve hb
			//remove id from list
			removeId, _ := strconv.Atoi(receiverId)

			//heartbeat failure
			hb := NewHeartbeat(id, memberList, FAILED)
			for i := 0; i < connections; i++ {
				receiverId := strconv.Itoa((id + i + 1) % len(memberList.Read()))
				receiverAddr := strings.Replace(host, idStr, receiverId, -1)  + ":" + strconv.Itoa(8000)
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
			membershipList := hb.GetMembershipList()
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

func UpdateMembershipLists(receivedList MembersList) {
	receivedNode := receivedList.GetHead()
	count := 0
	for count < len(receivedList.Read()) {
		//get node info from their membership list
		receivedStatus := receivedNode.GetStatus()
		receivedHbCount := receivedNode.GetHBCount()
		receivedId := receivedNode.GetId()

		currNode := memberList.GetNode(receivedId)

		if (currNode == nil && receivedStatus == ALIVE) {
			memberList.Insert(receivedNode)
			continue
		}

		// currStatus := currNode.GetStatus()
		currHBCount := currNode.GetHBCount()

		if(currHBCount < receivedHbCount) {
			if(receivedStatus == ALIVE) {
				r, rr, l, ll := receivedNode.GetNeighbors()
				currNode.SetHBCounter(receivedHbCount)
				currNode.SetNeighbors(r, rr, l, ll)
				currNode.SetStatus(receivedStatus)
			} else if (receivedStatus == LEAVE || receivedStatus == FAILED) {
				currNode.SetStatus(receivedStatus)
				go Cleanup(receivedId)
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
		//send heartbeat after certain duration
		time.After(heartbeatInterval)

		//increment heartbeat counter for node sending hb
		currNode := memberList.GetNode(id)
		currNode.IncrementHBCounter()

		hb := NewHeartbeat(id, memberList, ALIVE)
		SendOnce(hb, addr)
	}
}

func SendOnce(hb *Heartbeat, addr string) {
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
		fmt.Printf("entryMachineAddr %s", entryMachineAddr)
		SendOnce(entryHB, entryMachineAddr)

		//receive heartbeat from entry machine and update memberList
		ln, err := net.Listen("udp", entryMachineIdStr)
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
		receiverId := strconv.Itoa((id + i + 1) % len(memberList.Read()))
		receiverHost := strings.Replace(host, idStr, receiverId, -1)
		leave = false
		go Listen(8000 + i, receiverHost, receiverId)
		go Gossip(8000 + i, receiverHost, id)
	}
}

func Leave() {
	//remove self from membership list
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	memberList.Remove(id)

	// Send heartbeat to leave
	hb := NewHeartbeat(id, memberList, LEAVE)
	for i := 0; i < connections; i++ {
		receiverId := strconv.Itoa((id + i + 1) % len(memberList.Read()))
		receiverAddr := strings.Replace(host, idStr, receiverId, -1) + ":" + strconv.Itoa(8000)

		SendOnce(hb, receiverAddr)
	}

	// Kill goroutines for sending and recieving heartbeats
	leave = true
}