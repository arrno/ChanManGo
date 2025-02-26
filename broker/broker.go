package broker

import (
	"context"

	"github.com/arrno/chanban/domain"
)

type IBroker interface {
	CreateChannel(channel *domain.Channel) error
	DestroyChannel(channel *domain.Channel) error
	PublishMessage(ctx context.Context, channel *domain.Channel, m *domain.Message) error
	SubscribeForMessages(ctx context.Context, channel *domain.Channel, subscriber *domain.Subscription) error
	UnsubscribeForMessages(ctx context.Context, channel *domain.Channel, subscriber *domain.Subscription) error
}
