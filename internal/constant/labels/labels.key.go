package labels_constant

// LabelKey represents predefined keys for labels
const (
	// Work Type
	LabelRepeated = "REPEATED"
	LabelGroup    = "GROUP"
	LabelDaily    = "DAILY"

	// Status
	LabelPending    = "PENDING"
	LabelInProgress = "IN_PROGRESS"
	LabelCompleted  = "COMPLETED"
	LabelOverDue    = "OVER_DUE"
	LabelGiveUp     = "GIVE_UP"

	// Difficulty
	LabelDifficultyEasy   = "EASY"
	LabelDifficultyMedium = "MEDIUM"
	LabelDifficultyHard   = "HARD"

	// Priority
	LabelPriorityImportantUrgent       = "IMPORTANT_URGENT"
	LabelPriorityImportantNotUrgent    = "IMPORTANT_NOT_URGENT"
	LabelPriorityNotImportantUrgent    = "NOT_IMPORTANT_URGENT"
	LabelPriorityNotImportantNotUrgent = "NOT_IMPORTANT_NOT_URGENT"

	// Category
	LabelCategoryWork     = "WORK"
	LabelCategoryPersonal = "PERSONAL"
	LabelCategoryStudy    = "STUDY"
	LabelCategoryFamily   = "FAMILY"
	LabelCategoryFinance  = "FINANCE"
	LabelCategoryHealth   = "HEALTH"
	LabelCategorySocial   = "SOCIAL"
	LabelCategoryTravel   = "TRAVEL"

	// Draft
	LabelDraft = "DRAFT"
)
