package types

type ChangeDiff struct {
	ComponentReports []ComponentChangeDiff `json:"components"`
}

type ComponentChangeDiff struct {
	ComponentName string `json:"name"`
	OldVersion    string `json:"oldVersion"`
	NewVersion    string `json:"newVersion"`
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

func (c *ChangeDiff) Append(changeDiffs ...*ChangeDiff) {
	for _, diff := range changeDiffs {
		if diff != nil {
			c.ComponentReports = append(c.ComponentReports, diff.ComponentReports...)
		}
	}
}

func (c *ChangeDiff) Changed() bool {
	return len(c.ComponentReports) > 0
}
