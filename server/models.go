package mig

type User struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	WorkflowState string `json:"workflow_state"`
}
