/*
All Rights Reversed (ɔ)
*/

package output

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/redpierrot/bestow/internal/engine"
)

type mockSummary struct {
	counts   map[engine.ActionKind]int
	reverted int
}

func (ms mockSummary) Count(kind engine.ActionKind) int {
	return ms.counts[kind]
}

func (ms mockSummary) Reverted() int {
	return ms.reverted
}

type mockStatus struct {
	counts map[engine.State]int
}

func (ms mockStatus) Count(state engine.State) int {
	return ms.counts[state]
}

func TestOutput_printAction(t *testing.T) {
	tests := []struct {
		name   string
		action engine.ActionEvent
		label  string
		want   string
	}{
		{
			name: "success: no label",
			action: engine.ActionEvent{
				Action:    "link",
				Msg:       "dest -> src",
				EventType: engine.EventSuccess,
			},
			want: "   LINK: dest -> src\n",
		},
		{
			name: "success: with label",
			action: engine.ActionEvent{
				Action:    "link",
				Msg:       "dest -> src",
				EventType: engine.EventSuccess,
			},
			label: "[dryrun]",
			want:  "[dryrun]    LINK: dest -> src\n",
		},
		{
			name: "step: no label",
			action: engine.ActionEvent{
				Action:    "remove",
				Msg:       "dest",
				EventType: engine.EventStep,
			},
			want: " REMOVE: dest\n",
		},
		{
			name: "step: with label",
			action: engine.ActionEvent{
				Action:    "remove",
				Msg:       "dest",
				EventType: engine.EventStep,
			},
			label: "[dryrun]",
			want:  "[dryrun]  REMOVE: dest\n",
		},
		{
			name: "skip: no label",
			action: engine.ActionEvent{
				Action:    "skip",
				Msg:       "dest",
				EventType: engine.EventSkip,
			},
			want: "   SKIP: dest\n",
		},
		{
			name: "skip: with label",
			action: engine.ActionEvent{
				Action:    "skip",
				Msg:       "dest",
				EventType: engine.EventSkip,
			},
			label: "[dryrun]",
			want:  "[dryrun]    SKIP: dest\n",
		},
		{
			name: "undo: no label",
			action: engine.ActionEvent{
				Action:    "undo",
				Msg:       "dest",
				EventType: engine.EventUndo,
			},
			want: "   UNDO: dest\n",
		},
		{
			name: "undo: with label",
			action: engine.ActionEvent{
				Action:    "undo",
				Msg:       "dest",
				EventType: engine.EventUndo,
			},
			label: "[dryrun]",
			want:  "[dryrun]    UNDO: dest\n",
		},
		{
			name: "failure: no label",
			action: engine.ActionEvent{
				Action:    "create",
				Msg:       "dest",
				EventType: engine.EventFailure,
			},
			want: " CREATE: dest\n",
		},
		{
			name: "failure: with label",
			action: engine.ActionEvent{
				Action:    "create",
				Msg:       "dest",
				EventType: engine.EventFailure,
			},
			label: "[dryrun]",
			want:  "[dryrun]  CREATE: dest\n",
		},
		{
			name: "ignored",
			action: engine.ActionEvent{
				EventType: 1000,
			},
			want: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, err bytes.Buffer
			o := NewOutput(&out, &err)
			o.printAction(tc.action, tc.label)

			got := out.String()
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

type SummaryMock struct {
	counts   [1]int
	reverted int
}

func (s *SummaryMock) Count(action int) int {
	return s.counts[action]
}

func (s *SummaryMock) Reverted() int {
	return s.reverted
}

func TestOutput_PrintResult(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		result *engine.ExecuteResult
		want   string
	}{
		{
			name: "no result",
			want: "",
		},
		{
			name:  "quiet",
			level: Quiet,
			result: &engine.ExecuteResult{
				Events:  []engine.ActionEvent{},
				Summary: &engine.Summary{},
				DryRun:  false,
			},
		},
		{
			name: "normal",
			result: &engine.ExecuteResult{
				Events: []engine.ActionEvent{
					{
						Action:    "link",
						Msg:       "dest -> src",
						EventType: engine.EventSuccess,
					},
				},
				Summary: &engine.Summary{},
				DryRun:  false,
			},
			want: "   LINK: dest -> src\nno operations to execute\n",
		},
		{
			name: "normal: dryrun",
			result: &engine.ExecuteResult{
				Events: []engine.ActionEvent{
					{
						Action:    "link",
						Msg:       "dest -> src",
						EventType: engine.EventSuccess,
					},
				},
				Summary: &engine.Summary{},
				DryRun:  true,
			},
			want: "[dryrun]    LINK: dest -> src\nno operations to execute\n",
		},
		{
			name: "normal: nil summary",
			result: &engine.ExecuteResult{
				Events: []engine.ActionEvent{
					{
						Action:    "link",
						Msg:       "dest -> src",
						EventType: engine.EventSuccess,
					},
				},
				Summary: nil,
				DryRun:  true,
			},
			want: "[dryrun]    LINK: dest -> src\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, err bytes.Buffer
			o := NewOutput(&out, &err)
			o.SetLevel(tc.level)
			o.PrintResult(tc.result)
			got := out.String()
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOutput_printSummaryLine(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		summary summarizer
		want    string
	}{
		{
			name:  "multiple counts",
			level: Normal,
			summary: mockSummary{
				counts: map[engine.ActionKind]int{
					engine.ActionLink:  4,
					engine.ActionAdopt: 1,
				},
			},
			want: "link: 4   adopt: 1\n",
		},
		{
			name:  "with reverts",
			level: Normal,
			summary: mockSummary{
				counts: map[engine.ActionKind]int{
					engine.ActionLink:  4,
					engine.ActionAdopt: 1,
				},
				reverted: 5,
			},
			want: "link: 4   adopt: 1   reverted: 5\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, err bytes.Buffer
			o := NewOutput(&out, &err)
			o.SetLevel(tc.level)

			o.printSummaryLine(tc.summary)
			got := out.String()
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOutput_PrintStatus(t *testing.T) {
	ms := &mockStatus{
		counts: map[engine.State]int{
			engine.Stowed:   10,
			engine.Unstowed: 0,
		},
	}
	var out, err bytes.Buffer
	o := NewOutput(&out, &err)
	o.PrintStatus(ms)
	wantOut := "stowed: 10\n"
	wantErr := ""
	if wantOut != out.String() {
		t.Fatalf("got %q, want %s", &out, wantOut)
	}
	if wantErr != err.String() {
		t.Fatalf("got %q, want %s", &err, wantErr)
	}
}

func TestOutput_PrintCommandError(t *testing.T) {
	testErr := fmt.Errorf("sample error")
	var out, err bytes.Buffer
	o := NewOutput(&out, &err)
	o.PrintCommandError(testErr)
	wantErr := "sample error\n"
	wantOut := ""
	if err.String() != wantErr {
		t.Fatalf("got %q, want %q", &err, wantErr)
	}
	if out.String() != wantOut {
		t.Fatalf("got %q, want %q", &err, wantOut)
	}
}
