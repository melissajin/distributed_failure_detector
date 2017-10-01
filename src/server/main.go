package main

import (
	"time"
	"net"
	"strconv"
	"io/ioutil"
	"sync"
	"os"
	"regexp"
	"strings"
	"bufio"
	"fmt"
)

type Heartbeat struct {
	id int
	host string
	membershipList [] Machine
	timestamp time.Time
	status int
}

type Machine struct {
	id int
	timestamp time.Time
}

type MembersList struct {
	list [] Machine
	mu sync.Mutex
}

func (m *MembersList) Read() [] Machine{
	m.mu.Lock()
	list := m.list
	m.mu.Unlock()
	return list
}

func (m *MembersList) Insert(id int, timestamp time.Time) {
	m.mu.Lock()
	m.list = append(m.list, Machine{id, timestamp})
	m.mu.Unlock()
	timestamp = time.Now()
}

func (m *MembersList) GetIndex(id int) int{
	m.mu.Lock()
	idx := 0
	for i, v := range m.list {
		if(v.id == id) {
			idx = i
		}
	}
	m.mu.Unlock()
	return idx
}

func (m *MembersList) Copy(list [] Machine) {
	m.mu.Lock()
	m.list = list
	m.mu.Unlock()
	timestamp = time.Now()
}

func (m *MembersList) Remove(id int) {
	m.mu.Lock()
	for i, v := range m.list {
		if (v.id == id) {
			m.list[len(m.list) - 1], m.list[i] = m.list[i], m.list[len(m.list) - 1]
			m.list = m.list[:len(m.list)-1]
		}
	}
	m.mu.Unlock()
	timestamp = time.Now()
}

var memberList MembersList
var id int
var timestamp time.Time
var leave bool

const (
	connections = 4
	detectionTime = time.Second * 2		//seconds
	heartbeatInterval = time.Second * 1 //seconds
	entryMachineId = 1
)

// Status of machine, sent in heartbeat
const (
	ALIVE = iota	// 0
	JOIN			// 1
	LEAVE			// 2
)

//change ids to idx in membership list

func main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Command: ")
		text, _ := reader.ReadString('\n')
		if(strings.EqualFold(text, "join")) {
			Join()
		} else if(strings.EqualFold(text, "leave")) {
			Leave()
		} else if(strings.EqualFold(text, "list")) {
			fmt.Print(memberList.list)
		} else if(strings.EqualFold(text, "id")) {
			host,_ := os.Hostname()
			re, _ := regexp.Compile("-[0-9]+.")
			idStr := re.FindString(host)
			fmt.Print(idStr + "\n")
		} else {
			fmt.Println("Invalid Command")
		}
	}
}

func Listen(port int, receiverHost string, receiverId string) {
	addr := ":" + strconv.Itoa(port)
	ln, err := net.Listen("udp", addr)

	for {
		select {
		case leave == true:
			break
		case <- time.After(detectionTime):
			//remove id from list, heartbeat failure
			removeId, err := strconv.Atoi(receiverId)
			memberList.Remove(removeId)

		default:
			// accept and read heartbeat struct from server
			conn, err := ln.Accept()
			hb, _ := ioutil.ReadAll(conn)

			//extract id and membershiplist from struct
			id := hb[0]
			membershipList := hb[len(receiverHost):len(receiverHost) + len(memberList.list)]
			ts := hb[len(receiverHost) + 10:]

			//merge membership lists
			UpdateMembershipLists(id, membershipList, ts, ALIVE)

			conn.Close()
		}
	}
}

func UpdateMembershipLists(id int, membershipList []Machine, ts time.Time, status int) {
	if(status == ALIVE) {
		if (timestamp.Sub(ts) <= 0) {
			memberList.Copy(membershipList)
			timestamp = ts
		}
	} else if (status == JOIN) {
		memberList.Insert(id, ts)
	}
}

func GetIdentity() (string, int) {
	host, _ := os.Hostname()
	re, _ := regexp.Compile("-[0-9]+.")
	idStr := re.FindString(host)
	id, _ := strconv.Atoi(idStr)
	return host, id
}

func Send(port int, host string, id int) {
	addr := host + ":" + strconv.Itoa(port)
	for {
		select {
		case leave == true:
			break
		default:
			//send heartbeat after certain duration
			time.After(heartbeatInterval)
			hb := Heartbeat{id, host, memberList.Read(), timestamp, ALIVE}
			SendOnce(hb, addr)
		}
	}
}

func SendOnce(hb Heartbeat, addr string) {
	conn, _ := net.Dial("udp", addr)
	conn.Write([]byte(hb))
	conn.Close()
}

func Join() {
	host, id := GetIdentity()
	idStr := strconv.Itoa(id)
	entryMachineIdStr := strconv.Itoa(entryMachineId)
	timestamp = time.Now()

	//add self to membership list and contact fixed entry machine and get membership list
	memberList.Insert(id, timestamp)
	hb := Heartbeat{id, host, memberList.Read(), timestamp, JOIN}
	entryMachineAddr := strings.Replace(host, idStr, entryMachineIdStr, -1)
	SendOnce(hb, entryMachineAddr)

	// TODO: receive heartbeat from entry machine and update memberList

	// start 2 threads for each connection, each listening to different port
	for i := 0; i < connections; i++ {
		receiverId := strconv.Itoa((id + i + 1) % len(memberList.list))
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
	memberList.Remove(id)

	// Send heartbeat to leave
	hb := Heartbeat{id, host, memberList.Read(), time.Now(), LEAVE}
	for i := 0; i < connections; i++ {
		receiverId := strconv.Itoa((id + i + 1) % len(memberList.list))
		receiverAddr := strings.Replace(host, idStr, receiverId, -1)
		SendOnce(hb, receiverAddr)
	}

	// Kill goroutines for sending and recieving heartbeats
	leave = true
}