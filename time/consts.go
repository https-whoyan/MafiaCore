package time

// Time and deadline constants in seconds
const (
	VotingDeadline = 40

	RoleInfoCount = 10

	FakeVotingMinSeconds = 5
	FakeVotingMaxSeconds = 36

	LastWordDeadline = 60
)

// Everything below is automatically calculated

const (
	LastWordDeadlineMinutes = LastWordDeadline / 60
)
