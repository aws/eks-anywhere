package types

type ChangeReport struct {
	ComponentReports []ComponentChangeReport
}

type ComponentChangeReport struct {
	ComponentName string
	OldVersion    string
	NewVersion    string
}
