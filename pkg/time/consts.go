package time

// Time and deadline constants are described below
const (
	RegistrationDeadlineSeconds     = 300
	VotingGameConfigDeadlineSeconds = 300

	FakeVotingMinSeconds = 5
	FakeVotingMaxSeconds = 25
)

// Everything below is automatically calculated
const (
	RegistrationDeadlineMinutes     = RegistrationDeadlineSeconds / 60
	VotingGameConfigDeadlineMinutes = VotingGameConfigDeadlineSeconds / 60
)