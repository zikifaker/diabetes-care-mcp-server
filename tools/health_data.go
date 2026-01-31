package tools

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	bloodGlucoseRecordTableName = "blood_glucose_record"
	healthProfileTableName      = "health_profile"
	exerciseRecordTableName     = "exercise_record"
	defaultRecordsLimit         = 30
)

type BloodGlucoseRecord struct {
	Value        float32   `json:"value"`
	MeasuredAt   time.Time `json:"measuredAt"`
	DiningStatus string    `json:"diningStatus"`
}

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

type ExerciseRecord struct {
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Intensity   string    `json:"intensity"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Duration    int       `json:"duration"`
	PreGlucose  float32   `json:"pre_glucose"`
	PostGlucose float32   `json:"post_glucose"`
	Notes       string    `json:"notes"`
}

func FetchHealthData(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dataType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	email := ctx.Value("user_email").(string)

	switch dataType {
	case "blood_glucose":
		limit := req.GetInt("limit", defaultRecordsLimit)
		return mcp.NewToolResultJSON(getBloodGlucoseRecords(ctx, email, limit))

	case "health_profile":
		return mcp.NewToolResultJSON(getHealthProfile(ctx, email))

	case "exercise_records":
		limit := req.GetInt("limit", defaultRecordsLimit)
		return mcp.NewToolResultJSON(getExerciseRecords(ctx, email, limit))

	default:
		return mcp.NewToolResultError("invalid type param"), nil
	}
}

func getBloodGlucoseRecords(ctx context.Context, email string, limit int) []BloodGlucoseRecord {
	var records []BloodGlucoseRecord
	err := dao.DB.Table(bloodGlucoseRecordTableName).
		Select("value, measured_at, dining_status").
		Where("user_email = ?", email).
		Order("measured_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		slog.Error("Failed to get blood glucose records",
			"email", email,
			"err", err,
		)
		return nil
	}
	return records
}

func getHealthProfile(ctx context.Context, email string) *HealthProfile {
	var profile HealthProfile
	err := dao.DB.Table(healthProfileTableName).
		Select("gender, age, height, weight, dietary_preference, smoking_status, activity_level, diabetes_type, diagnosis_year, therapy_mode, medication, allergies, complications").
		Where("user_email = ?", email).
		First(&profile).Error
	if err != nil {
		slog.Error("Failed to get health profile",
			"email", email,
			"err", err,
		)
		return nil
	}
	return &profile
}

func getExerciseRecords(ctx context.Context, email string, limit int) []ExerciseRecord {
	var records []ExerciseRecord
	err := dao.DB.Table(exerciseRecordTableName).
		Select("type, name, intensity, start_at, end_at, duration, pre_glucose, post_glucose, notes").
		Where("user_email = ?", email).
		Order("measured_at DESC").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		slog.Error("Failed to get exercise records",
			"email", email,
			"err", err,
		)
		return nil
	}
	return records
}
