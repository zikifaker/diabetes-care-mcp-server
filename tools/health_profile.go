package tools

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

const healthProfileTableName = "health_profile"

type HealthProfile struct {
	Gender            string  `json:"gender"`
	Age               int     `json:"age"`
	Height            float32 `json:"height"`
	Weight            float32 `json:"weight"`
	DietaryPreference string  `json:"dietary_preference"`
	SmokingStatus     bool    `json:"smoking_status"`
	ActivityLevel     string  `json:"activity_level"`
	DiabetesType      string  `json:"diabetes_type"`
	DiagnosisYear     int     `json:"diagnosis_year"`
	TherapyMode       string  `json:"therapy_mode"`
	Medication        string  `json:"medication"`
	Allergies         string  `json:"allergies"`
	Complications     string  `json:"complications"`
}

func GetHealthProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	email := ctx.Value("user_email").(string)

	var profile HealthProfile
	err := dao.DB.Table(healthProfileTableName).
		Select("gender, age, height, weight, dietary_preference, smoking_status, activity_level, diabetes_type, diagnosis_year, therapy_mode, medication, allergies, complications").
		Where("user_email = ?", email).
		First(&profile).Error
	if err != nil {
		slog.Error("Failed to get health profile", "err", err)
		return nil, err
	}

	return mcp.NewToolResultJSON(profile)
}
