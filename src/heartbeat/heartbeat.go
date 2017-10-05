package heartbeat

import(
	. "membersList"
)

type Heartbeat struct {
	id int
	host string
	membershipList MembersList
	status int
}

func NewHeartbeat(id int, host string, membershipList MembersList, status int) *Heartbeat {
	return &Heartbeat{id, host, membershipList, status}
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