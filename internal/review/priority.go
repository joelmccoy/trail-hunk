package review

type Priority string

const (
	PriorityBlocker Priority = "blocker"
	PriorityHigh    Priority = "high"
	PriorityMedium  Priority = "medium"
	PriorityLow     Priority = "low"
	PriorityNote    Priority = "note"
)

type CommentCategory string

const (
	CategoryBug             CommentCategory = "bug"
	CategorySecurity        CommentCategory = "security"
	CategoryCorrectness     CommentCategory = "correctness"
	CategoryMaintainability CommentCategory = "maintainability"
	CategoryPerformance     CommentCategory = "performance"
	CategoryTest            CommentCategory = "test"
	CategoryQuestion        CommentCategory = "question"
)

type CommentStatus string

const (
	StatusSuggested CommentStatus = "suggested"
	StatusApproved  CommentStatus = "approved"
	StatusEdited    CommentStatus = "edited"
	StatusDismissed CommentStatus = "dismissed"
	StatusSubmitted CommentStatus = "submitted"
)

type CommentSource string

const (
	SourceAI   CommentSource = "ai"
	SourceUser CommentSource = "user"
)
