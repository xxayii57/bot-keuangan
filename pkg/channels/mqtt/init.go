package mqtt

import (
	"github.com/xxayii57/bot-keuangan/pkg/bus"
	"github.com/xxayii57/bot-keuangan/pkg/channels"
	"github.com/xxayii57/bot-keuangan/pkg/config"
)

func init() {
	channels.RegisterSafeFactory(
		config.ChannelMQTT,
		func(bc *config.Channel, cfg *config.MQTTSettings, b *bus.MessageBus) (channels.Channel, error) {
			return NewMQTTChannel(bc, cfg, b)
		},
	)
}
