package vk

import (
	"github.com/xxayii57/intimclaw/pkg/bus"
	"github.com/xxayii57/intimclaw/pkg/channels"
	"github.com/xxayii57/intimclaw/pkg/config"
)

func init() {
	channels.RegisterFactory(
		config.ChannelVK,
		func(channelName, channelType string, cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
			bc := cfg.Channels[channelName]
			if bc == nil {
				return nil, channels.ErrSendFailed
			}
			return NewVKChannel(channelName, bc, b)
		},
	)
}
