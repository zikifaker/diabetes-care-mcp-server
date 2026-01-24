package tools

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

type HealthProfile struct {
	DiabetesType  string `json:"diabetes_type"`
	Medication    string `json:"medication"`
	Complications string `json:"complications"`
}

func GetHealthProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	email := ctx.Value("user_email").(string)

	var profile HealthProfile
	err := dao.DB.Table("health_profile").
		Select("diabetes_type, medication, complications").
		Where("user_email = ?", email).
		First(&profile).Error
	if err != nil {
		slog.Error("Failed to get health profile", "err", err)
		return nil, err
	}

	return mcp.NewToolResultJSON(profile)
}
