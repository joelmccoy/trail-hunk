package review

type ReviewSession struct {
	Plan     WalkthroughPlan
	Cursor   ReviewCursor
	Comments []ReviewComment
	nextID   int
}

type WalkthroughPlan struct {
	Overview    string
	Risks       []Risk
	ReviewOrder []ReviewStep
}

type Risk struct {
	Priority Priority
	Category CommentCategory
	Summary  string
}

type ReviewStep struct {
	ID          string
	FilePath    string
	HunkID      string
	Title       string
	Summary     string
	Why         string
	Focus       []string
	Suggestions []ReviewComment
}

type ReviewCursor struct {
	StepIndex int
	FileIndex int
	HunkIndex int
}

func (s *ReviewSession) NextStep() bool {
	if s.Cursor.StepIndex+1 >= len(s.Plan.ReviewOrder) {
		return false
	}
	s.Cursor.StepIndex++
	return true
}

func (s *ReviewSession) PreviousStep() bool {
	if s.Cursor.StepIndex == 0 {
		return false
	}
	s.Cursor.StepIndex--
	return true
}
