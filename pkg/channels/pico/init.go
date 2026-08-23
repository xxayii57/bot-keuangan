package pico

import (
	"github.com/xxayii57/bot-keuangan/pkg/bus"
	"github.com/xxayii57/bot-keuangan/pkg/channels"
	"github.com/xxayii57/bot-keuangan/pkg/config"
)

func init() {
	channels.RegisterFactory(
		config.ChannelPico,
		func(channelName, channelType string, cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
			bc := cfg.Channels[channelName]
			decoded, err := bc.GetDecoded()
			if err != nil {
				return nil, err
			}
			c, ok := decoded.(*config.PicoSettings)
			if !ok {
				return nil, channels.ErrSendFailed
			}
			ch, err := NewPicoChannel(bc, c, b)
			if err != nil {
				return nil, err
			}
			if channelName != config.ChannelPico {
				ch.SetName(channelName)
			}
			return ch, nil
		},
	)
	channels.RegisterFactory(
		config.ChannelPicoClient,
		func(channelName, channelType string, cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
			bc := cfg.Channels[channelName]
			decoded, err := bc.GetDecoded()
			if err != nil {
				return nil, err
			}
			c, ok := decoded.(*config.PicoClientSettings)
			if !ok {
				return nil, channels.ErrSendFailed
			}
			ch, err := NewPicoClientChannel(bc, c, b)
			if err != nil {
				return nil, err
			}
			if channelName != config.ChannelPicoClient {
				ch.SetName(channelName)
			}
			return ch, nil
		},
	)
}
