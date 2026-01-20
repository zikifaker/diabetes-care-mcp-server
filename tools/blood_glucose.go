package tools

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const defaultGetBloodGlucoseRecordsLimit = 30

type BloodGlucoseRecord struct {
	Value        float32   `json:"value"`
	MeasuredAt   time.Time `json:"measuredAt"`
	DiningStatus string    `json:"diningStatus"`
}

func GetBloodGlucoseRecords(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := req.GetInt("limit", defaultGetBloodGlucoseRecordsLimit)

	var records []BloodGlucoseRecord
	err := dao.DB.Table("blood_glucose_record").
		Select("value, measured_at, dining_status").
		Order("measured_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultJSON(records)
}
