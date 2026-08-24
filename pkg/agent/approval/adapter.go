package approval

import (
	"context"

	"github.com/xxayii57/bot-keuangan/pkg/agent"
)

var _ agent.ToolApprover = (*Gate)(nil)

// ApproveTool implements agent.ToolApprover by delegating to the Telegram gate.
func (g *Gate) ApproveTool(ctx context.Context, req *agent.ToolApprovalRequest) (agent.ApprovalDecision, error) {
	if g == nil || req == nil {
		return agent.ApprovalDecision{Approved: true}, nil
	}

	if !g.NeedsApproval(req.Tool, req.Arguments) {
		return agent.ApprovalDecision{Approved: true}, nil
	}

	chatID := ""
	if req.Context != nil && req.Context.Inbound != nil {
		chatID = req.Context.Inbound.ChatID
	}

	approved, err := g.PromptAndWait(ctx, req.Tool, req.Arguments, chatID)
	if err != nil {
		return agent.ApprovalDecision{}, err
	}
	if !approved {
		return agent.ApprovalDecision{Approved: false, Reason: "operator rejected in Telegram"}, nil
	}
	return agent.ApprovalDecision{Approved: true}, nil
}
