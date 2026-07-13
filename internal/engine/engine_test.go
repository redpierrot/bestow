/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"os"
	"slices"
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
		name        string
		fs          *mockFileSystem
		actions     []fileAction
		events      []ActionEvent
		wantEvents  []ActionEvent
		wantSummary Summary
		wantErr     bool
		wantErrIs   error
	}{
		{
			name: "one event undo: skip",
			actions: []fileAction{
				newFileActionSkip("src", "dest", "reason", newTestLogger()),
			},
			events: []ActionEvent{
				{
					Action:    actionSkip,
					Msg:       "src -> dest [reason]",
					EventType: EventSkip,
				},
			},
			wantEvents: []ActionEvent{
				{
					Action:    actionSkip,
					Msg:       "src -> dest [reason]",
					EventType: EventSkip,
				},
				{
					EventType: EventIgnore,
				},
			},
		},
		{
			name: "one event undo: link",
			actions: []fileAction{
				newFileActionLink("src", "dest", newTestLogger()),
			},
			fs: &mockFileSystem{},
			events: []ActionEvent{
				{
					Action:    actionLink,
					Msg:       "src -> dest",
					EventType: EventSuccess,
				},
			},
			wantEvents: []ActionEvent{
				{
					Action:    actionLink,
					Msg:       "src -> dest",
					EventType: EventSuccess,
				},
				{
					Action:    actionRemove,
					Msg:       "dest",
					EventType: EventUndo,
				},
			},
			wantSummary: Summary{
				reverted: 1,
			},
		},
		{
			name: "multiple events",
			actions: []fileAction{
				newFileActionLink("src1", "dest1", newTestLogger()),
				newFileActionLink("src2", "dest2", newTestLogger()),
				newFileActionLink("src3", "dest3", newTestLogger()),
			},
			fs: &mockFileSystem{},
			events: []ActionEvent{
				{
					Action:    actionLink,
					Msg:       "src1 -> dest1",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src2 -> dest2",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src3 -> dest3",
					EventType: EventSuccess,
				},
			},
			wantEvents: []ActionEvent{
				{
					Action:    actionLink,
					Msg:       "src1 -> dest1",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src2 -> dest2",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src3 -> dest3",
					EventType: EventSuccess,
				},
				{
					Action:    actionRemove,
					Msg:       "dest3",
					EventType: EventUndo,
				},
				{
					Action:    actionRemove,
					Msg:       "dest2",
					EventType: EventUndo,
				},
				{
					Action:    actionRemove,
					Msg:       "dest1",
					EventType: EventUndo,
				},
			},
			wantSummary: Summary{
				reverted: 3,
			},
		},
		{
			name: "multiple events - undo fail in middle",
			actions: []fileAction{
				newFileActionLink("src1", "dest1", newTestLogger()),
				newFileActionLink("src2", "dest2", newTestLogger()),
				newFileActionLink("src3", "dest3", newTestLogger()),
			},
			fs: &mockFileSystem{
				removeFn: func(path string) error {
					if path == "dest2" {
						return os.ErrPermission
					}
					return nil
				},
			},
			events: []ActionEvent{
				{
					Action:    actionLink,
					Msg:       "src1 -> dest1",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src2 -> dest2",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src3 -> dest3",
					EventType: EventSuccess,
				},
			},
			wantEvents: []ActionEvent{
				{
					Action:    actionLink,
					Msg:       "src1 -> dest1",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src2 -> dest2",
					EventType: EventSuccess,
				},
				{
					Action:    actionLink,
					Msg:       "src3 -> dest3",
					EventType: EventSuccess,
				},
				{
					Action:    actionRemove,
					Msg:       "dest3",
					EventType: EventUndo,
				},
			},
			wantSummary: Summary{
				reverted: 1,
			},
			wantErr:   true,
			wantErrIs: os.ErrPermission,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := Summary{}
			e := newTestEngine(tc.fs, nil)
			result, err := e.undoFileActions(tc.actions, &summary, tc.events)
			validateErrScenario(t, tc.wantErr, err, tc.wantErrIs)
			if !slices.Equal(result.Events, tc.wantEvents) {
				t.Fatalf("got %v, want %v", result.Events, tc.wantEvents)
			}
			if summary != tc.wantSummary {
				t.Fatalf("got %v, want %v", summary, tc.wantSummary)
			}
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
			want:    Summary{counts: [numActionKinds]int{ActionUpToDate: 1}},
			action:  newFileActionUpToDate("", "", "", newTestLogger()),
		},
		{
			name:    "skip",
			summary: Summary{},
			want:    Summary{counts: [numActionKinds]int{ActionSkip: 1}},
			action:  newFileActionSkip("", "", "", newTestLogger()),
		},
		{
			name:    "link",
			summary: Summary{},
			want:    Summary{counts: [numActionKinds]int{ActionLink: 1}},
			action:  newFileActionLink("", "", newTestLogger()),
		},
		{
			name:    "replace",
			summary: Summary{},
			want:    Summary{counts: [numActionKinds]int{ActionReplace: 1}},
			action:  newFileActionReplace("", "", newTestLogger()),
		},
		{
			name:    "backup",
			summary: Summary{},
			want:    Summary{counts: [numActionKinds]int{ActionBackup: 1}},
			action:  newFileActionBackup("", "", "", newTestLogger()),
		},
		{
			name:    "adopt",
			summary: Summary{},
			want:    Summary{counts: [numActionKinds]int{ActionAdopt: 1}},
			action:  newFileActionAdopt("", "", newTestLogger()),
		},
		{
			name:    "remove",
			summary: Summary{},
			want:    Summary{counts: [numActionKinds]int{ActionRemove: 1}},
			action:  newFileActionRemove("", "", newTestLogger()),
		},
		{
			name:    "undo",
			summary: Summary{},
			isUndo:  true,
			want:    Summary{reverted: 1},
			action:  newFileActionRemove("", "", newTestLogger()),
		},
		{
			name:    "undo skip",
			summary: Summary{},
			isUndo:  true,
			want:    Summary{},
			action:  newFileActionSkip("", "", "", newTestLogger()),
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
