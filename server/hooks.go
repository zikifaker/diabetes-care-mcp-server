package server

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const methodToolCompleted = "tool_completed"

func pushCallToolResult(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
	mcpServer := server.ServerFromContext(ctx)
	if mcpServer != nil {
		err := mcpServer.SendNotificationToClient(ctx, methodToolCompleted, map[string]any{
			"tool":   message.Params.Name,
			"result": result.Content,
		})
		if err != nil {
			slog.Error("error sending notification",
				"tool", message.Params.Name,
				"err", err,
			)
		}
	}
}
