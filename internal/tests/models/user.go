package models

type TestRenameUserProvider struct{}

var TestRenameUserProviderInstance = &TestRenameUserProvider{}

func (rP *TestRenameUserProvider) RenameUser(_, _, _ string) error { return nil }
