package roles

type Role struct {
	Name           string `json:"name" bson:"name" db:"name" yaml:"name" xml:"name" xlsx:"mame"`
	Team           Team   `json:"team" bson:"team" db:"team" yaml:"team" xml:"team" xlsx:"mame"`
	NightVoteOrder int    `json:"nightVoteOrder" bson:"nightVoteOrder" db:"nightVoteOrder" yaml:"nightVoteOrder" xml:"nightVoteOrder" xls:"nightVoteOrder"`
	// Presents whether to execute immediately, the action of the role.
	UrgentCalculation bool
	// Allows for calculations to be made in the correct order after night.
	CalculationOrder int
	// Presents whether 2 player IDs are used in night actions of the role at once.
	IsTwoVotes  bool
	Description string
}

var MappedRoles = map[string]*Role{
	"Citizen":   Citizen,
	"Detective": Detective,
	"Doctor":    Doctor,
	"Don":       Don,
	"Fool":      Fool,
	"Mafia":     Mafia,
	"Maniac":    Maniac,
	"Peaceful":  Peaceful,
	"Whore":     Whore,
}
