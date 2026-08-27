package feishu

import (
	"github.com/xxayii57/intimclaw/pkg/bus"
	"github.com/xxayii57/intimclaw/pkg/channels"
	"github.com/xxayii57/intimclaw/pkg/config"
)

func init() {
	channels.RegisterFactory(
		config.ChannelFeishu,
		func(channelName, channelType string, cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
			bc := cfg.Channels[channelName]
			decoded, err := bc.GetDecoded()
			if err != nil {
				return nil, err
			}
			c, ok := decoded.(*config.FeishuSettings)
			if !ok {
				return nil, channels.ErrSendFailed
			}
			return NewFeishuChannel(bc, c, b)
		},
	)
}
