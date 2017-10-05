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
	id int
	membershipList MembersList
	status int
}

func NewHeartbeat(id int, membershipList MembersList, status int) *Heartbeat {
	return &Heartbeat{id, membershipList, status}
}

func (hb *Heartbeat) GetId() int {
	return hb.id
}

func (hb *Heartbeat) GetMembershipList() MembersList {
	return hb.membershipList
}

func (hb *Heartbeat) GetStatus() int {
	return hb.status
}