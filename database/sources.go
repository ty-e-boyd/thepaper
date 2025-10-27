package database

import (
	"fmt"
)

// CreateSource creates a new RSS feed source in the database
func CreateSource(name, category, url string, active bool) (*Source, error) {
	source := &Source{
		Name:     name,
		Category: category,
		URL:      url,
		Active:   active,
	}

	result := DB.Create(source)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create source: %w", result.Error)
	}

	return source, nil
}

// GetAllActiveSources returns all active RSS feed sources
func GetAllActiveSources() ([]Source, error) {
	var sources []Source
	result := DB.Where("active = ?", true).Find(&sources)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active sources: %w", result.Error)
	}
	return sources, nil
}

// GetAllSources returns all RSS feed sources (active and inactive)
func GetAllSources() ([]Source, error) {
	var sources []Source
	result := DB.Find(&sources)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get sources: %w", result.Error)
	}
	return sources, nil
}

// GetSourcesByCategory returns all active sources for a specific category
func GetSourcesByCategory(category string) ([]Source, error) {
	var sources []Source
	result := DB.Where("category = ? AND active = ?", category, true).Find(&sources)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get sources by category: %w", result.Error)
	}
	return sources, nil
}

// GetSourceByURL finds a source by its URL
func GetSourceByURL(url string) (*Source, error) {
	var source Source
	result := DB.Where("url = ?", url).First(&source)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find source: %w", result.Error)
	}
	return &source, nil
}

// UpdateSourceActive updates a source's active status
func UpdateSourceActive(sourceID uint, active bool) error {
	result := DB.Model(&Source{}).Where("id = ?", sourceID).Update("active", active)
	if result.Error != nil {
		return fmt.Errorf("failed to update source active status: %w", result.Error)
	}
	return nil
}

// DeleteSource soft deletes a source
func DeleteSource(sourceID uint) error {
	result := DB.Delete(&Source{}, sourceID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete source: %w", result.Error)
	}
	return nil
}
