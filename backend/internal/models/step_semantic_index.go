package models

import "time"

type StepSemanticIndex struct {
	StepID       uint      `gorm:"primaryKey;column:step_id"`
	ProcessID    uint      `gorm:"column:process_id;index"`
	StepType     StepType  `gorm:"column:step_type;type:varchar(30)"`
	StepName     string    `gorm:"column:step_name;type:text;index"`
	EmbeddingJSON string   `gorm:"column:embedding_json;type:text"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (StepSemanticIndex) TableName() string {
	return "step_semantic_index"
}

type StepSuggestion struct {
	StepID      uint    `json:"stepId"`
	ProcessID   uint    `json:"processId"`
	ProcessName string  `json:"processName"`
	StepName    string  `json:"stepName"`
	StepType    StepType `json:"stepType"`
	Score       float64 `json:"score"`
}

