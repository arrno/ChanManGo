package domain

type Subscription struct {
	Id        string                      `json:"id"`
	ClientId  string                      `json:"clientId"`
	TopicName string                      `json:"topicName"`
	State     STATE                       `json:"state"`
	Archived  bool                        `json:"archived"`
	Callback  func(message Message) error `json:"-"`
}

func NewSubcription() *Subscription {
	return &Subscription{}
}
