package types

type ChangeDiff struct {
	ComponentReports []ComponentChangeDiff
}

type ComponentChangeDiff struct {
	ComponentName string
	OldVersion    string
	NewVersion    string
}
