package types

type ChangeDiff struct {
	ComponentReports []ComponentChangeDiff
}

type ComponentChangeDiff struct {
	ComponentName string
	OldVersion    string
	NewVersion    string
}

func NewChangeDiff(componentReports ...*ComponentChangeDiff) *ChangeDiff {
	reports := make([]ComponentChangeDiff, 0, len(componentReports))
	for _, r := range componentReports {
		if r != nil {
			reports = append(reports, *r)
		}
	}

	return &ChangeDiff{
		ComponentReports: reports,
	}
}
