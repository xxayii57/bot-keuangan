package channels

import (
	"context"
	"testing"

	"github.com/xxayii57/bot-keuangan/pkg/commands"
)

type mockRegistrar struct{}

func (mockRegistrar) RegisterCommands(context.Context, []commands.Definition) error { return nil }

func TestCommandRegistrarCapable_Compiles(t *testing.T) {
	var _ CommandRegistrarCapable = mockRegistrar{}
}
