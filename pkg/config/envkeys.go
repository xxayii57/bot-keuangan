// IntimClaw - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 IntimClaw contributors

package config

import (
	"os"
	"path/filepath"

	"github.com/xxayii57/bot-keuangan/pkg"
)

// Runtime environment variable keys for the intimclaw process.
// These control the location of files and binaries at runtime and are read
// directly via os.Getenv / os.LookupEnv. All intimclaw-specific keys use the
// INTIMCLAW_ prefix. Reference these constants instead of inline string
// literals to keep all supported knobs visible in one place and to prevent
// typos.
const (
	// EnvHome overrides the base directory for all intimclaw data
	// (config, workspace, skills, auth store, …).
	// Default: ~/.intimclaw
	EnvHome = "INTIMCLAW_HOME"

	// EnvConfig overrides the full path to the JSON config file.
	// Default: $INTIMCLAW_HOME/config.json
	EnvConfig = "INTIMCLAW_CONFIG"

	// EnvBuiltinSkills overrides the directory from which built-in
	// skills are loaded.
	// Default: <cwd>/skills
	EnvBuiltinSkills = "INTIMCLAW_BUILTIN_SKILLS"

	// EnvBinary overrides the path to the intimclaw executable.
	// Used by the web launcher when spawning the gateway subprocess.
	// Default: resolved from the same directory as the current executable.
	EnvBinary = "INTIMCLAW_BINARY"

	// EnvGatewayHost overrides the host address for the gateway server.
	// Default: "localhost"
	EnvGatewayHost = "INTIMCLAW_GATEWAY_HOST"
)

func GetHome() string {
	homePath, _ := os.UserHomeDir()
	if intimclawHome := os.Getenv(EnvHome); intimclawHome != "" {
		homePath = intimclawHome
	} else if homePath != "" {
		homePath = filepath.Join(homePath, pkg.DefaultIntimClawHome)
	}
	if homePath == "" {
		homePath = "."
	}
	return homePath
}
