package models

// ---------------------------------------------------------------------------
// Sessions — V13 API: /api/v1/sessions
// Used for tracking async operations and job run history.
// ---------------------------------------------------------------------------

// SessionModel is the full V13 session response.
// Note: The client/async.go has a simpler SessionModel for polling.
// This is the complete model for data source use.
type FullSessionModel struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	JobID           string         `json:"jobId"`
	SessionType     string         `json:"sessionType"`
	CreationTime    string         `json:"creationTime"`
	EndTime         string         `json:"endTime,omitempty"`
	State           ESessionState  `json:"state"`
	ProgressPercent int            `json:"progressPercent,omitempty"`
	Result          *SessionResult `json:"result,omitempty"`
	ResourceID      string         `json:"resourceId,omitempty"`
	ResourceRef     string         `json:"resourceReference,omitempty"`
	ParentSessionID string         `json:"parentSessionId,omitempty"`
	USN             int            `json:"usn"`
}

// SessionResult holds the result of a completed session.
type SessionResult struct {
	Result     ESessionResult `json:"result"`
	Message    string         `json:"message,omitempty"`
	IsCanceled bool           `json:"isCanceled,omitempty"`
}

// SessionLogRecord represents a log entry within a session.
type SessionLogRecord struct {
	ID          int    `json:"id,omitempty"`
	Status      string `json:"status,omitempty"`
	StartTime   string `json:"startTime,omitempty"`
	UpdateTime  string `json:"updateTime,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}
