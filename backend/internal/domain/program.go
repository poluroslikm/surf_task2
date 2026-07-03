package domain

// ProgramDifficulty — slots/models.yaml#/components/schemas/ProgramDifficulty.
type ProgramDifficulty string

const (
	DifficultyNovice      ProgramDifficulty = "novice"
	DifficultyExperienced ProgramDifficulty = "experienced"
)

// Program is a read-only catalog entry (data-model.md → Program). No capacity field here —
// seat limits live on Slot, not on Program (see migrations/00001_init.sql comment).
type Program struct {
	ID          string
	Name        string
	Difficulty  ProgramDifficulty
	PhotoURL    string
	Ingredients []string
	Allergens   []string
}
