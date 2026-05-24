package agent

type AgentResponse struct {
	Summary             string   `json:"summary"`
	KeyFindings         []string `json:"key_findings"`
	Risks               []string `json:"risks"`
	RecommendedNextSteps []string `json:"recommended_next_steps"`
	UsedContextFields   []string `json:"used_context_fields"`
}
