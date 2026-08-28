package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/xxayii57/intimclaw/pkg/config"
)

// setkeyCommand handles "/setkey" to configure API keys for models
func setkeyCommand() Definition {
	return Definition{
		Name:        "setkey",
		Description: "Set API key and optional base URL for a model",
		Usage:       "/setkey <model_name> <api_key> [api_base]",
		Handler:     setkeyHandler(),
	}
}

func setkeyHandler() Handler {
	return func(_ context.Context, req Request, rt *Runtime) error {
		if rt == nil {
			return req.Reply(unavailableMsg)
		}
		if req.Reply == nil {
			return nil
		}

		args := req.Args
		if len(args) < 2 {
			return req.Reply("⚠️ *Format salah!*\n\nGunakan:\n`/setkey <nama_model> <api_key> [api_base]`\n\nContoh:\n`/setkey deepseek sk-xxxx https://9router.intim.my.id/v1`")
		}

		modelName := args[0]
		apiKey := args[1]
		apiBase := ""
		if len(args) >= 3 {
			apiBase = args[2]
		}

		p := filepath.Join(config.GetHome(), "config.json")
		cfg2, err := config.LoadConfig(p)
		if err != nil {
			return req.Reply(fmt.Sprintf("❌ *Gagal load config:* %v", err))
		}

		found := false
		for _, m := range cfg2.ModelList {
			if m != nil && m.ModelName == modelName {
				m.APIKeys = config.SimpleSecureStrings(apiKey)
				if apiBase != "" {
					m.APIBase = apiBase
				}
				m.Enabled = true
				found = true
				break
			}
		}

		if !found {
			return req.Reply(fmt.Sprintf("❌ *Model `%s` tidak ditemukan di config!*", modelName))
		}

		if err := config.SaveConfig(p, cfg2); err != nil {
			return req.Reply(fmt.Sprintf("❌ *Gagal menyimpan config:* %v", err))
		}

		msgText := fmt.Sprintf("✅ *Config Model `%s` Berhasil Diupdate!*\n\nKey: `••••••••`", modelName)
		if apiBase != "" {
			msgText += fmt.Sprintf("\nBase URL: `%s`", apiBase)
		}
		return req.Reply(msgText)
	}
}
