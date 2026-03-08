package models

import (
	"time"

	"gorm.io/gorm"
)

type StepType string

const (
	StepStart        StepType = "START"
	StepEnd          StepType = "END"
	StepIntermediate StepType = "INTERMEDIATE"
	StepSubprocess   StepType = "SUBPROCESS"
	StepOperation    StepType = "OPERATION"
	StepCondition    StepType = "CONDITION"
)

type ProcessFolder struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	ParentID *uint
	Parent   *ProcessFolder  `gorm:"foreignKey:ParentID"`
	Children []ProcessFolder `gorm:"foreignKey:ParentID"`

	Processes []Process `gorm:"foreignKey:FolderID"`

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-" swaggerignore:"true"`
}

type Process struct {
	ID uint `gorm:"primaryKey"`

	Name        string `gorm:"not null"`
	Description string
	Regulations string

	OwnerID uint
	Owner   Employee `gorm:"foreignKey:OwnerID"`

	FolderID *uint
	Folder   *ProcessFolder

	IsActive bool `gorm:"default:true"`

	Versions []ProcessVersion `gorm:"foreignKey:ProcessID"`

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"  json:"-" swaggerignore:"true"`
}

type ProcessVersion struct {
	ID uint `gorm:"primaryKey"`

	ProcessID uint
	Process   Process `gorm:"foreignKey:ProcessID"`

	Version int

	IsPublished bool

	Steps []ProcessStep `gorm:"foreignKey:ProcessVersionID"`

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-" swaggerignore:"true"`
}

type ProcessStep struct {
	ID uint `gorm:"primaryKey"`

	ProcessVersionID uint
	ProcessVersion   ProcessVersion `gorm:"foreignKey:ProcessVersionID"`

	StepOrder int
	Name      string

	Type StepType `gorm:"type:varchar(30)"`

	Description string

	SubprocessID *uint
	Subprocess   *Process `gorm:"foreignKey:SubprocessID"`

	Executors []Employee `gorm:"many2many:process_step_executors"`

	Metrics *StepMetrics `gorm:"foreignKey:StepID"`

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-" swaggerignore:"true"`
}

type StepMetrics struct {
	ID uint `gorm:"primaryKey"`

	StepID uint

	PlannedTimeMin int
	Step           ProcessStep `gorm:"foreignKey:StepID"`

	TimeStatistics *StepTimeStatistics `gorm:"foreignKey:MetricsID"`
}

type StepTimeStatistics struct {
	ID uint `gorm:"primaryKey"`

	MetricsID uint
	Metrics   StepMetrics `gorm:"foreignKey:MetricsID"`

	MinTime    int
	MinPercent float64

	AvgTime    int
	AvgPercent float64

	MaxTime    int
	MaxPercent float64

	WeightedAvg float64
}

type StepMeasurement struct {
	ID uint `gorm:"primaryKey"`

	StepID uint
	Step   ProcessStep

	MeasurementNumber int `gorm:"check:measurement_number <= 3"`

	StartedAt  *time.Time
	FinishedAt *time.Time

	PausedSeconds int
	DurationSec   int

	Pauses []MeasurementPause `gorm:"foreignKey:MeasurementID"`
}

type MeasurementPause struct {
	ID uint `gorm:"primaryKey"`

	MeasurementID uint
	Measurement   StepMeasurement `gorm:"foreignKey:MeasurementID"`

	PauseStart time.Time
	PauseEnd   *time.Time
}

type ProcessStepLink struct {
	ID uint `gorm:"primaryKey"`

	FromStepID uint
	FromStep   ProcessStep `gorm:"foreignKey:FromStepID"`

	ToStepID uint
	ToStep   ProcessStep `gorm:"foreignKey:ToStepID"`

	ConditionText string
}

type ProcessRegistryFolder struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	ParentID *uint  `json:"parentId"`

	Processes []ProcessShortDTO        `json:"processes"`
	Children  []*ProcessRegistryFolder `json:"children"`
}

type ProcessShortDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CreateProcessRequest struct {
	Name     string `json:"name" binding:"required"`
	FolderID *uint  `json:"folderId,omitempty"`
}

type CreateFolderRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parentId,omitempty"`
}

type CreateVersionRequest struct {
	ProcessID uint `json:"processId" binding:"required"`
}
