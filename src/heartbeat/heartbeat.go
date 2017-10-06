package heartbeat

import(
	. "membersList"
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

