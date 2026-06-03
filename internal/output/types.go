/*
All Rights Reversed (ɔ)
*/

package output

type Type int

const (
	TypeSuccess Type = iota
	TypeStep
	TypeWarn
)

type Summary struct {
	Stowed   int
	Unstowed int
	Replaced int
	Backed   int
	Adopted  int
	Skipped  int
	UpToDate int
}
