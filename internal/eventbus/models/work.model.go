package event_models

type WorkMessage struct {
	Name                string   `json:"name"`
	ShortDescriptions   string   `json:"short_descriptions"`
	DetailedDescription string   `json:"detailed_description"`
	StartDate           string   `json:"start_date"`
	EndDate             string   `json:"end_date"`
	DifficultyKey       string   `json:"difficulty_key"`
	PriorityKey         string   `json:"priority_key"`
	CategoryKey         string   `json:"category_key"`
	SubTasks            []string `json:"sub_tasks"`
}
