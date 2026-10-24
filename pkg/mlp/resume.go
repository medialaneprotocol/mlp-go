package mlp

// ResumeRejectReason describes why a resume was rejected.
type ResumeRejectReason string

const (
	ResumeOK        ResumeRejectReason = "ok"
	ResumeRejected  ResumeRejectReason = "rejected"
	ResumeExpired   ResumeRejectReason = "expired"
)

// ResumeAction is an action taken during resume.
type ResumeAction struct {
	Kind    string
	Tokens  []InputToken
	Pending uint64
}

// InputToken represents a held input (key or pointer).
type InputToken struct {
	Kind string
	Code string
}

// InputTokenKey creates a key token.
func InputTokenKey(code string) InputToken { return InputToken{Kind: "key", Code: code} }
