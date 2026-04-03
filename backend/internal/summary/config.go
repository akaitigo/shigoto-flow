package summary

type SummaryConfig struct {
	FocusTopics   []string `json:"focus_topics"`
	ExcludeTopics []string `json:"exclude_topics"`
	MaxLength     int      `json:"max_length"`
	DetailLevel   string   `json:"detail_level"`
}

func DefaultConfig() SummaryConfig {
	return SummaryConfig{
		MaxLength:   1500,
		DetailLevel: "standard",
	}
}
