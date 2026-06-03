/*
All Rights Reversed (ɔ)
*/

package engine

type ExecuteSummary struct {
	Actions          []ActionEvent
	OperationSummary *Summary
	Dryrun           bool
}

type Summary struct {
	Stowed   int
	Unstowed int
	Replaced int
	BackedUp int
	Adopted  int
	Skipped  int
	UpToDate int
}
