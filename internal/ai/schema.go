package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var allowedPriorities = map[string]struct{}{
	"blocker": {},
	"high":    {},
	"medium":  {},
	"low":     {},
	"note":    {},
}

var allowedCategories = map[string]struct{}{
	"bug":             {},
	"security":        {},
	"correctness":     {},
	"maintainability": {},
	"performance":     {},
	"test":            {},
	"question":        {},
}

func DecodeReviewResponse(raw []byte) (ReviewResponse, error) {
	var response ReviewResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return ReviewResponse{}, fmt.Errorf("decode review response JSON: %w", err)
	}
	if err := ValidateReviewResponse(response); err != nil {
		return ReviewResponse{}, err
	}
	return response, nil
}

func ValidateReviewResponse(response ReviewResponse) error {
	if strings.TrimSpace(response.Overview) == "" {
		return errors.New("review response is missing overview")
	}

	for i, risk := range response.Risks {
		if err := validatePriority(risk.Priority); err != nil {
			return fmt.Errorf("risk %d: %w", i, err)
		}
		if risk.Category != "" {
			if err := validateCategory(risk.Category); err != nil {
				return fmt.Errorf("risk %d: %w", i, err)
			}
		}
	}

	for i, step := range response.ReviewOrder {
		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("review step %d is missing id", i)
		}
		if strings.TrimSpace(step.Summary) == "" {
			return fmt.Errorf("review step %q is missing summary", step.ID)
		}
		if strings.TrimSpace(step.Why) == "" {
			return fmt.Errorf("review step %q is missing why", step.ID)
		}
		for j, suggestion := range step.Suggestions {
			if strings.TrimSpace(suggestion.Body) == "" {
				return fmt.Errorf("review step %q suggestion %d is missing body", step.ID, j)
			}
			if err := validatePriority(suggestion.Priority); err != nil {
				return fmt.Errorf("review step %q suggestion %d: %w", step.ID, j, err)
			}
			if err := validateCategory(suggestion.Category); err != nil {
				return fmt.Errorf("review step %q suggestion %d: %w", step.ID, j, err)
			}
		}
	}

	return nil
}

func validatePriority(priority string) error {
	if _, ok := allowedPriorities[priority]; !ok {
		return fmt.Errorf("invalid priority %q", priority)
	}
	return nil
}

func validateCategory(category string) error {
	if _, ok := allowedCategories[category]; !ok {
		return fmt.Errorf("invalid category %q", category)
	}
	return nil
}
