package review

type ReviewSession struct {
	Repo     RepoRef
	PR       PullRequest
	Plan     WalkthroughPlan
	Cursor   ReviewCursor
	Comments []ReviewComment
	nextID   int
}

type RepoRef struct {
	Owner  string
	Name   string
	Root   string
	Branch string
}

type PullRequest struct {
	Number int
	Title  string
	Body   string
	State  string
	URL    string
}

type StartupContext struct {
	Repo    RepoRef
	PR      *PullRequest
	Message string
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
	GroupID     string
	GroupTitle  string
	LayerIndex  int
	LayerTitle  string
	Summary     string
	Why         string
	Focus       []string
	DiffLines   []DiffLine
	Suggestions []ReviewComment
}

type DiffLineKind string

const (
	DiffLineContext DiffLineKind = "context"
	DiffLineAdded   DiffLineKind = "added"
	DiffLineDeleted DiffLineKind = "deleted"
)

type DiffLine struct {
	Kind    DiffLineKind
	OldLine *int
	NewLine *int
	Text    string
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
