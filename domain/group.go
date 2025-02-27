package domain

type Group struct {
	Id       string
	OwnerId  string
	PortalId string
	Name     string
	state    STATE
}

type ClientGroup struct {
	Id       string
	ClientId string
	GroupId  string
	state    STATE
}

type ChannelGroup struct {
	Id        string
	ChannelId string
	GroupId   string
	state     STATE
}
