/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"
)

type subTest struct {
	name      string
	fs        *mockFileSystem
	want      []ActionEvent
	wantErr   bool
	wantErrIs error
}

func TestFileAction_execute(t *testing.T) {
	tests := []struct {
		name   string
		action func(l *slog.Logger) fileAction
		cases  []subTest
	}{
		{
			name: "up-to-date",
			action: func(l *slog.Logger) fileAction {
				return newFileActionUpToDate("src", "dest", "reason", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{EventType: EventIgnore},
					},
				},
			},
		},
		{
			name: "skip",
			action: func(l *slog.Logger) fileAction {
				return newFileActionSkip("src", "dest", "reason", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpSkip,
							Msg:       "src -> dest [reason]",
							EventType: EventSkip,
						},
					},
				},
			},
		},
		{
			name: "link",
			action: func(l *slog.Logger) fileAction {
				return newFileActionLink("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpLink,
							Msg:       "dest -> src",
							EventType: EventSuccess,
						},
					},
				},
				{
					name: "fail",
					fs: &mockFileSystem{
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "replace",
			action: func(l *slog.Logger) fileAction {
				return newFileActionReplace("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventStep,
						},
						{
							Action:    fileOpLink,
							Msg:       "dest -> src",
							EventType: EventSuccess,
						},
						{
							Action:    fileOpRemove,
							Msg:       "dest.bestow.tmp",
							EventType: EventIgnore,
						},
					},
				},
				{
					name: "move fail",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "link fail - move pass",
					fs: &mockFileSystem{
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "link fail - recovery fail",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							if strings.Contains(src, "tmp") {
								return os.ErrPermission
							}
							return nil
						},
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      nil,
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					// TODO: Should validate the warning
					name: "tmp remove fail",
					fs: &mockFileSystem{
						removeFn: func(path string) error {
							if strings.Contains(path, "tmp") {
								return os.ErrPermission
							}
							return nil
						},
					},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventStep,
						},
						{
							Action:    fileOpLink,
							Msg:       "dest -> src",
							EventType: EventSuccess,
						},
						{
							Action:    fileOpLeftover,
							Msg:       "temp file dest.bestow.tmp",
							EventType: EventFailure,
						},
					},
				},
			},
		},
		{
			name: "backup",
			action: func(l *slog.Logger) fileAction {
				return newFileActionBackup("src", "dest", "dest.bestow.backup", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpBackup,
							Msg:       "dest -> dest.bestow.backup",
							EventType: EventStep,
						},
						{
							Action:    fileOpLink,
							Msg:       "dest -> src",
							EventType: EventSuccess,
						},
					},
				},
				{
					name: "move fails",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "link fail - recovery fail",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							if strings.Contains(src, "backup") {
								return os.ErrPermission
							}
							return nil
						},
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want: []ActionEvent{
						{
							Action:    fileOpBackup,
							Msg:       "dest -> dest.bestow.backup",
							EventType: EventStep,
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "link fail - recovery pass",
					fs: &mockFileSystem{
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "adopt",
			action: func(l *slog.Logger) fileAction {
				return newFileActionAdopt("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpAdopt,
							Msg:       "dest -> src",
							EventType: EventStep,
						},
						{
							Action:    fileOpLink,
							Msg:       "dest -> src",
							EventType: EventSuccess,
						},
					},
				},
				{
					name: "move fails",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "link fail - recovery fail",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							if strings.Contains(src, "src") {
								return os.ErrPermission
							}
							return nil
						},
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want: []ActionEvent{
						{
							Action:    fileOpAdopt,
							Msg:       "dest -> src",
							EventType: EventStep,
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "link fail - recovery pass",
					fs: &mockFileSystem{
						linkFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "remove",
			action: func(l *slog.Logger) fileAction {
				return newFileActionRemove("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventSuccess,
						},
					},
				},
				{
					name: "remove fail",
					fs: &mockFileSystem{
						removeFn: func(path string) error {
							return os.ErrPermission
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fa := tc.action(newTestLogger())
			for _, st := range tc.cases {
				t.Run(st.name, func(t *testing.T) {
					events, err := fa.execute(st.fs)
					if st.want != nil {
						if !slices.Equal(events, st.want) {
							t.Fatalf("got %v, want %v", events, st.want)
						}
					}
					validateErrScenario(t, st.wantErr, err, st.wantErrIs)
				})
			}
		})
	}
}

func TestFileAction_undo(t *testing.T) {
	tests := []struct {
		name   string
		action func(l *slog.Logger) fileAction
		cases  []subTest
	}{
		{
			name: "up-to-date",
			action: func(l *slog.Logger) fileAction {
				return newFileActionUpToDate("src", "dest", "reason", l)
			},
			cases: []subTest{
				{
					name: "run",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{EventType: EventIgnore},
					},
				},
			},
		},
		{
			name: "skip",
			action: func(l *slog.Logger) fileAction {
				return newFileActionSkip("src", "dest", "reason", l)
			},
			cases: []subTest{
				{
					name: "run",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{EventType: EventIgnore},
					},
				},
			},
		},
		{
			name: "link",
			action: func(l *slog.Logger) fileAction {
				return newFileActionLink("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "link file",
					fs: &mockFileSystem{
						removeFn: func(path string) error {
							return nil
						},
					},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventUndo,
						},
					},
				},
				{
					name: "remove fail",
					fs: &mockFileSystem{
						removeFn: func(path string) error {
							return os.ErrPermission
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "replace",
			action: func(l *slog.Logger) fileAction {
				return newFileActionReplace("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventUndo,
						},
					},
				},
				{
					name: "remove fail",
					fs: &mockFileSystem{
						removeFn: func(path string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "backup",
			action: func(l *slog.Logger) fileAction {
				return newFileActionBackup("src", "dest", "dest.bestow.backup", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpRestore,
							Msg:       "dest.bestow.backup -> dest",
							EventType: EventUndo,
						},
					},
				},
				{
					name: "move fails",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "adopt",
			action: func(l *slog.Logger) fileAction {
				return newFileActionAdopt("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventUndo,
						},
						{
							Action:    fileOpRestore,
							Msg:       "src -> dest",
							EventType: EventUndo,
						},
					},
				},
				{
					name: "remove fails",
					fs: &mockFileSystem{
						removeFn: func(path string) error {
							return os.ErrPermission
						},
					},
					want:      []ActionEvent{},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
				{
					name: "move fails",
					fs: &mockFileSystem{
						moveFn: func(src, target string) error {
							return os.ErrPermission
						},
					},
					want: []ActionEvent{
						{
							Action:    fileOpRemove,
							Msg:       "dest",
							EventType: EventUndo,
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
		{
			name: "remove",
			action: func(l *slog.Logger) fileAction {
				return newFileActionRemove("src", "dest", l)
			},
			cases: []subTest{
				{
					name: "no errors",
					fs:   &mockFileSystem{},
					want: []ActionEvent{
						{
							Action:    fileOpLink,
							Msg:       "dest -> src",
							EventType: EventUndo,
						},
					},
				},
				{
					name: "link fail",
					fs: &mockFileSystem{
						linkFn: func(src, dest string) error {
							return os.ErrPermission
						},
					},
					wantErr:   true,
					wantErrIs: os.ErrPermission,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fa := tc.action(newTestLogger())
			for _, st := range tc.cases {
				t.Run(st.name, func(t *testing.T) {
					events, err := fa.undo(st.fs)
					if st.want != nil {
						if !slices.Equal(events, st.want) {
							t.Fatalf("got %v, want %v", events, st.want)
						}
					}
					validateErrScenario(t, st.wantErr, err, st.wantErrIs)
				})
			}
		})
	}
}
