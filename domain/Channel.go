package domain

type Channel struct {
	Id        string `json:"id"`
	ClientId  string `json:"clientId"`
	TopicName string `json:"topicName"`
	State     STATE  `json:"state"`
	Archived  bool   `json:"archived"`
}
