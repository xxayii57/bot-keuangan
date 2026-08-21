// intimclaw.ic — Built by xxayii — IntimClaw Discord Bot v0.1.0
package agent

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	token string
	agent *Agent
	dg    *discordgo.Session
}

func NewDiscordBot(token string, agent *Agent) *DiscordBot {
	return &DiscordBot{
		token: token,
		agent: agent,
	}
}

func (b *DiscordBot) Start() error {
	var err error
	b.dg, err = discordgo.New("Bot " + b.token)
	if err != nil {
		return fmt.Errorf("failed to create discord session: %w", err)
	}

	b.dg.AddHandler(b.messageCreate)
	b.dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	err = b.dg.Open()
	if err != nil {
		return fmt.Errorf("failed to open discord connection: %w", err)
	}

	fmt.Println("[intimclaw] Discord bot started successfully.")
	return nil
}

func (b *DiscordBot) Stop() {
	if b.dg != nil {
		b.dg.Close()
	}
}

func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	isMentioned := false
	for _, u := range m.Mentions {
		if u.ID == s.State.User.ID {
			isMentioned = true
			break
		}
	}

	isDM := m.GuildID == ""

	if !isMentioned && !isDM {
		return
	}

	prompt := m.Content
	prompt = strings.ReplaceAll(prompt, fmt.Sprintf("<@!%s>", s.State.User.ID), "")
	prompt = strings.ReplaceAll(prompt, fmt.Sprintf("<@%s>", s.State.User.ID), "")
	prompt = strings.TrimSpace(prompt)

	if prompt == "" {
		return
	}

	s.ChannelTyping(m.ChannelID)

	resp, err := b.agent.Run(prompt)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error: %v", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, resp)
}
