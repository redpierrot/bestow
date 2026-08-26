/*
All Rights Reversed (t *testing.T)
*/

package engine

import (
	"slices"
	"testing"
)

func TestActionUnstow_cleanup(t *testing.T) {
	l := newTestLogger()
	tests := []struct {
		name       string
		cmd        *UnstowCommand
		candidates []operationCandidate
		root       string
		want       []fileAction
	}{
		{
			name: "remove parents",
			cmd:  &UnstowCommand{},
			candidates: []operationCandidate{
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/init.lua", "/Users/ru/.config/nvim/plugins/init.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/yazi.lua", "/Users/ru/.config/nvim/plugins/yazi/yazi.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/keys.lua", "/Users/ru/.config/nvim/plugins/yazi/keys.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/init.lua", "/Users/ru/.config/nvim/init.lua"),
			},
			root: "/Users/ru/",
			want: []fileAction{
				newFileActionRemoveDir("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/yazi.lua", "/Users/ru/.config/nvim/plugins/yazi", l),
				newFileActionRemoveDir("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/keys.lua", "/Users/ru/.config/nvim/plugins", l),
				newFileActionRemoveDir("/Users/ru/dotfiles/nvim/.config/nvim/plugins/init.lua", "/Users/ru/.config/nvim", l),
				newFileActionRemoveDir("/Users/ru/dotfiles/nvim/.config/nvim/init.lua", "/Users/ru/.config", l),
			},
		},
		{
			name: "keep parents",
			cmd:  &UnstowCommand{KeepEmptyParents: true},
			candidates: []operationCandidate{
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/init.lua", "/Users/ru/.config/nvim/plugins/init.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/yazi.lua", "/Users/ru/.config/nvim/plugins/yazi/yazi.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/keys.lua", "/Users/ru/.config/nvim/plugins/yazi/keys.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/init.lua", "/Users/ru/.config/nvim/init.lua"),
			},
			root: "/Users/ru/",
			want: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cmd.cleanup(tc.candidates, l, tc.root)
			if !slices.EqualFunc(got, tc.want, equalFileAction) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestActionUnstow_getDirs(t *testing.T) {
	tests := []struct {
		name       string
		candidates []operationCandidate
		root       string
		want       []string
	}{
		{
			name: "paths with same parent directory",
			candidates: []operationCandidate{
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/init.lua", "/Users/ru/.config/nvim/plugins/init.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/yazi.lua", "/Users/ru/.config/nvim/plugins/yazi/yazi.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/keys.lua", "/Users/ru/.config/nvim/plugins/yazi/keys.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/init.lua", "/Users/ru/.config/nvim/init.lua"),
			},
			root: "/Users/ru/",
			want: []string{
				"/Users/ru/.config/nvim/plugins/yazi",
				"/Users/ru/.config/nvim/plugins",
				"/Users/ru/.config/nvim",
				"/Users/ru/.config",
			},
		},
		{
			name: "paths outside root",
			candidates: []operationCandidate{
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/init.lua", "/Users/ru/.config/nvim/plugins/init.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/yazi.lua", "/Users/ru/.config/nvim/plugins/yazi/yazi.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/keys.lua", "/Users/ru/.config/nvim/plugins/yazi/keys.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/init.lua", "/Users/ru/.config/nvim/init.lua"),
			},
			root: "/Users/thisaru/",
			want: []string{},
		},
		{
			name: "non rel paths",
			candidates: []operationCandidate{
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/init.lua", "/Users/ru/.config/nvim/plugins/init.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/yazi.lua", "/Users/ru/.config/nvim/plugins/yazi/yazi.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/plugins/yazi/keys.lua", "/Users/ru/.config/nvim/plugins/yazi/keys.lua"),
				candidate("/Users/ru/dotfiles/nvim/.config/nvim/init.lua", "/Users/ru/.config/nvim/init.lua"),
			},
			root: "/tmp/sandbox",
			want: []string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getDirs(tc.candidates, tc.root)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
