package commands

import (
	"context"
	"fmt"
	"strings"
)

func sessionsCommand() Definition {
	return Definition{
		Name:        "sessions",
		Description: "Manage chat sessions",
		SubCommands: []SubCommand{
			{
				Name:        "list",
				Description: "List all saved sessions",
				Handler:     sessionsListHandler(),
			},
			{
				Name:        "switch",
				Description: "Switch to a different session",
				ArgsUsage:   "<session-id>",
				Handler:     sessionsSwitchHandler(),
			},
			{
				Name:        "info",
				Description: "Show current session details",
				Handler:     sessionsInfoHandler(),
			},
		},
	}
}

func sessionsListHandler() Handler {
	return func(_ context.Context, req Request, rt *Runtime) error {
		if rt == nil || rt.ListSessions == nil {
			return req.Reply(unavailableMsg)
		}
		sessions := rt.ListSessions()
		if len(sessions) == 0 {
			return req.Reply("No sessions found.")
		}
		var b strings.Builder
		b.WriteString("📋 Saved Sessions:\n\n")
		for i, s := range sessions {
			marker := "  "
			b.WriteString(fmt.Sprintf("%s%d) %s\n", marker, i+1, s))
		}
		b.WriteString("\nCurrent: " + req.ChatID)
		return req.Reply(b.String())
	}
}

func sessionsSwitchHandler() Handler {
	return func(_ context.Context, req Request, rt *Runtime) error {
		if rt == nil || rt.ListSessions == nil {
			return req.Reply(unavailableMsg)
		}
		target := nthToken(req.Text, 2)
		if target == "" {
			return req.Reply("Usage: /sessions switch <session-id>\n\nUse /sessions list to see available sessions.")
		}
		sessions := rt.ListSessions()
		found := false
		for _, s := range sessions {
			if s == target || fmt.Sprintf("%d", indexOf(sessions, s)+1) == target {
				target = s
				found = true
				break
			}
		}
		if !found {
			return req.Reply(fmt.Sprintf("Session %q not found. Use /sessions list to see available sessions.", target))
		}
		return req.Reply(fmt.Sprintf("🔄 Session: %s\nNote: To switch session, start a new conversation.", target))
	}
}

func sessionsInfoHandler() Handler {
	return func(_ context.Context, req Request, rt *Runtime) error {
		return req.Reply(fmt.Sprintf(
			"📌 Current Session\n\nKey: %s\nChannel: %s\n\nUse /context to see token usage and /clear to reset this session.",
			req.ChatID, req.Channel))
	}
}

func indexOf(ss []string, s string) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return -1
}
