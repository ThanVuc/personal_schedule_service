package models

type GenerationWorksModel struct {
	UserID            string `bson:"user_id" json:"user_id"`
	Prompts           string `bson:"prompts" json:"prompts"`
	LocalDate         string `bson:"local_date" json:"local_date"`
	AdditionalContext string `bson:"additional_context" json:"additional_context"`
	Constraints       string `bson:"constraints" json:"constraints"`
	UserPersonality   string `bson:"user_personality" json:"user_personality"`
}

type TimeRange struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}
