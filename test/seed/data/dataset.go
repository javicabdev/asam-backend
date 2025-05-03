package data

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// DatasetType represents the type of dataset to generate
type DatasetType string

const (
	// MinimalType represents a minimal dataset for basic testing
	MinimalType DatasetType = "minimal"

	// FullType represents a complete dataset for comprehensive testing
	FullType DatasetType = "full"

	// ScenarioType represents a scenario-specific dataset
	ScenarioType DatasetType = "scenario"
)

// Dataset represents a dataset factory
func Dataset(db *sqlx.DB, seeder Seeder, datasetType DatasetType, options ...string) Seedable {
	switch datasetType {
	case MinimalType:
		return NewMinimalDataset(db, seeder)
	case FullType:
		return NewFullDataset(db, seeder)
	case ScenarioType:
		// The first option is the scenario name
		scenarioName := "payment_overdue" // Default scenario
		if len(options) > 0 {
			scenarioName = options[0]
		}
		return NewScenarioDataset(db, seeder, scenarioName)
	default:
		// Default to minimal dataset
		return NewMinimalDataset(db, seeder)
	}
}

// Seedable represents a type that can seed data
type Seedable interface {
	Seed(ctx context.Context) error
}

// Add convenience function for minimal dataset - renamed to avoid conflict
func NewMinimalDatasetHelper(seeder Seeder) Seedable {
	if s, ok := seeder.(interface{ GetDB() *sqlx.DB }); ok {
		return NewMinimalDataset(s.GetDB(), seeder)
	}
	return nil
}

// Add convenience function for full dataset - renamed to avoid conflict
func NewFullDatasetHelper(seeder Seeder) Seedable {
	if s, ok := seeder.(interface{ GetDB() *sqlx.DB }); ok {
		return NewFullDataset(s.GetDB(), seeder)
	}
	return nil
}

// Add convenience function for scenario dataset - renamed to avoid conflict
func NewScenarioDatasetHelper(seeder Seeder, name string) Seedable {
	if s, ok := seeder.(interface{ GetDB() *sqlx.DB }); ok {
		return NewScenarioDataset(s.GetDB(), seeder, name)
	}
	return nil
}
