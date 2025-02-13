package model

type Group struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (group Group) GetId() int {
	return group.Id
}

func (group *Group) SetId(id int) {
	group.Id = id
}

type Participant struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	GroupId int    `json:"groupId"`
}

func (participant Participant) GetId() int {
	return participant.Id
}

func (participant *Participant) SetId(id int) {
	participant.Id = id
}
