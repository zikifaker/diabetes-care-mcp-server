package tools

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	bloodGlucoseRecordTableName        = "blood_glucose_record"
	defaultGetBloodGlucoseRecordsLimit = 30
)

type BloodGlucoseRecord struct {
	Value        float32   `json:"value"`
	MeasuredAt   time.Time `json:"measuredAt"`
	DiningStatus string    `json:"diningStatus"`
}

func GetBloodGlucoseRecords(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := req.GetInt("limit", defaultGetBloodGlucoseRecordsLimit)
	email := ctx.Value("user_email").(string)

	var records []BloodGlucoseRecord
	err := dao.DB.Table(bloodGlucoseRecordTableName).
		Select("value, measured_at, dining_status").
		Where("user_email = ?", email).
		Order("measured_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		slog.Error("Failed to get blood glucose records", "err", err)
		return nil, err
	}

	return mcp.NewToolResultJSON(records)
}
