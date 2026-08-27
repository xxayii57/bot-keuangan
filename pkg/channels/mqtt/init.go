package mqtt

import (
	"github.com/xxayii57/intimclaw/pkg/bus"
	"github.com/xxayii57/intimclaw/pkg/channels"
	"github.com/xxayii57/intimclaw/pkg/config"
)

func init() {
	channels.RegisterSafeFactory(
		config.ChannelMQTT,
		func(bc *config.Channel, cfg *config.MQTTSettings, b *bus.MessageBus) (channels.Channel, error) {
			return NewMQTTChannel(bc, cfg, b)
		},
	)
}
