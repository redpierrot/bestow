/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"testing"
)

func TestEngine_NewEngine(t *testing.T) {
}

func TestEngine_Execute(t *testing.T) {
}

func TestEngine_executeFileActions(t *testing.T) {
}

func TestEngine_undoFileActions(t *testing.T) {
	tests := []struct {
		name      string
		input     *Summary
		want      *Summary
		action    fileAction
		isUndo    bool
		wantErr   bool
		wantErrIs error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestEngine_updateSummary(t *testing.T) {
	tests := []struct {
		name    string
		summary Summary
		want    Summary
		action  fileAction
		isUndo  bool
	}{
		{
			name:    "up-to-date",
			summary: Summary{},
			want: Summary{
				UpToDate: 1,
			},
			action: newFileActionUpToDate("", "", "", newTestLogger()),
		},
		{
			name:    "skip",
			summary: Summary{},
			want: Summary{
				Skipped: 1,
			},
			action: newFileActionSkip("", "", "", newTestLogger()),
		},
		{
			name:    "link",
			summary: Summary{},
			want: Summary{
				Stowed: 1,
			},
			action: newFileActionLink("", "", newTestLogger()),
		},
		{
			name:    "replace",
			summary: Summary{},
			want: Summary{
				Replaced: 1,
			},
			action: newFileActionReplace("", "", newTestLogger()),
		},
		{
			name:    "backup",
			summary: Summary{},
			want: Summary{
				BackedUp: 1,
			},
			action: newFileActionBackup("", "", "", newTestLogger()),
		},
		{
			name:    "adopt",
			summary: Summary{},
			want: Summary{
				Adopted: 1,
			},
			action: newFileActionAdopt("", "", newTestLogger()),
		},
		{
			name:    "remove",
			summary: Summary{},
			want: Summary{
				Unstowed: 1,
			},
			action: newFileActionRemove("", "", newTestLogger()),
		},
		{
			name:    "undo",
			summary: Summary{},
			isUndo:  true,
			want: Summary{
				Reverted: 1,
			},
			action: newFileActionRemove("", "", newTestLogger()),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEngine(&mockFileSystem{}, nil)
			e.updateSummary(tc.action, &tc.summary, tc.isUndo)
			if tc.summary != tc.want {
				t.Fatalf("got %v, want %v", tc.summary, tc.want)
			}
		})
	}
}
