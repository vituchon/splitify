package model

type Group struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Participant struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	GroupId int    `json:"groupId"`
}
