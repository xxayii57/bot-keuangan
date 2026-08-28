package commands

import (
	"context"
	"fmt"
	"strings"
)

// modelCommand handles "/model" and displays a list of available models
// with an inline keyboard for quick switching in Telegram.
func modelCommand() Definition {
	return Definition{
		Name:        "model",
		Description: "Show available models and pick one",
		Usage:       "/model",
		Handler:     modelListHandler(),
	}
}

func modelListHandler() Handler {
	return func(_ context.Context, req Request, rt *Runtime) error {
		if rt == nil {
			return req.Reply(unavailableMsg)
		}
		if req.Reply == nil {
			return nil
		}

		// List configured models
		var b strings.Builder
		b.WriteString("🤖 *Available Models*\n\n")

		current := ""
		if rt.GetModelInfo != nil {
			if name, _ := rt.GetModelInfo(); name != "" {
				current = name
			}
		}

		var models []string
		if rt.Config != nil {
			for _, m := range rt.Config.ModelList {
				if m == nil || !m.Enabled {
					continue
				}
				models = append(models, m.ModelName)
			}
		}

		if len(models) == 0 {
			b.WriteString("No models configured. Use:\n/model add <model-name>\n")
		} else {
			for i, m := range models {
				marker := "  "
				if m == current {
					marker = "→"
				}
				b.WriteString(fmt.Sprintf("%s `%d)` %s\n", marker, i+1, m))
			}
		}
		b.WriteString("\nChat this bot to switch model.")
		return req.Reply(b.String())
	}
}
