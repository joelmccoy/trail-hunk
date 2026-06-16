package review

import (
	"errors"
	"fmt"
	"strings"
)

type ReviewComment struct {
	ID        string
	FilePath  string
	Side      string
	Line      int
	StartLine *int
	Body      string
	Priority  Priority
	Category  CommentCategory
	Status    CommentStatus
	Source    CommentSource
}

func (s *ReviewSession) AcceptSuggestion(id string) error {
	comment, err := s.findComment(id)
	if err != nil {
		return err
	}
	if comment.Status != StatusSuggested {
		return fmt.Errorf("comment %q is not suggested", id)
	}
	comment.Status = StatusApproved
	return nil
}

func (s *ReviewSession) DismissSuggestion(id string) error {
	comment, err := s.findComment(id)
	if err != nil {
		return err
	}
	if comment.Status != StatusSuggested {
		return fmt.Errorf("comment %q is not suggested", id)
	}
	comment.Status = StatusDismissed
	return nil
}

func (s *ReviewSession) EditComment(id string, body string) error {
	if strings.TrimSpace(body) == "" {
		return errors.New("comment body cannot be empty")
	}

	comment, err := s.findComment(id)
	if err != nil {
		return err
	}
	if comment.Status == StatusDismissed || comment.Status == StatusSubmitted {
		return fmt.Errorf("comment %q cannot be edited from status %q", id, comment.Status)
	}

	comment.Body = body
	comment.Status = StatusEdited
	return nil
}

func (s *ReviewSession) AddManualComment(comment ReviewComment) ReviewComment {
	s.nextID++
	comment.ID = fmt.Sprintf("user-%d", s.nextID)
	comment.Source = SourceUser
	comment.Status = StatusApproved
	s.Comments = append(s.Comments, comment)
	return comment
}

func (s ReviewSession) ApprovedComments() []ReviewComment {
	var approved []ReviewComment
	for _, comment := range s.Comments {
		if comment.Status == StatusApproved || comment.Status == StatusEdited {
			approved = append(approved, comment)
		}
	}
	return approved
}

func (s *ReviewSession) MarkApprovedSubmitted() {
	for i := range s.Comments {
		if s.Comments[i].Status == StatusApproved || s.Comments[i].Status == StatusEdited {
			s.Comments[i].Status = StatusSubmitted
		}
	}
	for stepIndex := range s.Plan.ReviewOrder {
		for suggestionIndex := range s.Plan.ReviewOrder[stepIndex].Suggestions {
			status := s.Plan.ReviewOrder[stepIndex].Suggestions[suggestionIndex].Status
			if status == StatusApproved || status == StatusEdited {
				s.Plan.ReviewOrder[stepIndex].Suggestions[suggestionIndex].Status = StatusSubmitted
			}
		}
	}
}

func (s *ReviewSession) findComment(id string) (*ReviewComment, error) {
	for i := range s.Comments {
		if s.Comments[i].ID == id {
			return &s.Comments[i], nil
		}
	}
	return nil, fmt.Errorf("comment %q not found", id)
}
