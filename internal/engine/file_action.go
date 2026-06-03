/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"

	"github.com/ThisaruGuruge/bestow/internal/file"
)

const (
	actionLink    = "link"
	actionBackup  = "backup"
	actionSkip    = "skip"
	actionAdopt   = "adopt"
	actionRemove  = "remove"
	actionCreated = "created"
)

type EventType int

const (
	EventSuccess EventType = iota
	EventStep
	EventSkip
	EventWarn
	EventIgnore
)

type ActionType int

const (
	UpToDate ActionType = iota
	Skip
	Link
	Replace
	Backup
	Adopt
	Remove
)

func (a ActionType) String() string {
	switch a {
	case UpToDate:
		return "up-to-date"
	case Skip:
		return "skip"
	case Link:
		return "link"
	case Replace:
		return "replace"
	case Backup:
		return "backup"
	case Adopt:
		return "adopt"
	case Remove:
		return "remove"
	default:
		return fmt.Sprintf("Unknown %d", a)
	}
}

type ActionEvent struct {
	Action    string
	Msg       string
	EventType EventType
}

const backupExtension = ".bestow.backup"

type FileAction interface {
	Execute(fs file.System) ([]ActionEvent, error)
	Type() ActionType
	Source() string
	Destination() string
}

type fileActionBase struct {
	source      string
	destination string
	logger      *slog.Logger
}

func (fab fileActionBase) Source() string {
	return fab.source
}

func (fab fileActionBase) Destination() string {
	return fab.destination
}

type FileActionUpToDate struct {
	fileActionBase
	reason string
}

func newFileActionUpToDate(source, destination, reason string, l *slog.Logger) *FileActionUpToDate {
	return &FileActionUpToDate{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
		reason: reason,
	}
}

func (f *FileActionUpToDate) Execute(fs file.System) ([]ActionEvent, error) {
	return []ActionEvent{
		{EventType: EventIgnore},
	}, nil
}

func (f *FileActionUpToDate) Type() ActionType {
	return UpToDate
}

type FileActionSkip struct {
	fileActionBase
	reason string
}

func newFileActionSkip(source, destination, reason string, l *slog.Logger) *FileActionSkip {
	return &FileActionSkip{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
		reason: reason,
	}
}

func (f *FileActionSkip) Execute(fs file.System) ([]ActionEvent, error) {
	return []ActionEvent{
		{
			Action:    actionSkip,
			Msg:       fmt.Sprintf("%s -> %s [%s]", f.source, f.destination, f.reason),
			EventType: EventSkip,
		},
	}, nil
}

func (f *FileActionSkip) Type() ActionType {
	return Skip
}

type FileActionLink struct {
	fileActionBase
}

func newFileActionLink(source, destination string, l *slog.Logger) *FileActionLink {
	return &FileActionLink{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *FileActionLink) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Link(f.source, f.destination); err != nil {
		return nil, err
	}
	return []ActionEvent{
		{
			Action:    actionLink,
			Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
			EventType: EventSuccess,
		},
	}, nil
}

func (f *FileActionLink) Type() ActionType {
	return Link
}

type FileActionReplace struct {
	fileActionBase
}

func newFileActionReplace(source, destination string, l *slog.Logger) *FileActionReplace {
	return &FileActionReplace{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *FileActionReplace) Execute(fs file.System) ([]ActionEvent, error) {
	var events []ActionEvent
	tmp := f.destination + backupExtension
	if err := fs.Move(f.destination, tmp); err != nil {
		return nil, err
	}
	removeStep := ActionEvent{
		Action:    actionRemove,
		Msg:       f.destination,
		EventType: EventStep,
	}
	events = append(events, removeStep)
	if err := fs.Link(f.source, f.destination); err != nil {
		if err := fs.Move(tmp, f.destination); err != nil {
			f.logger.Warn("failed to restore the tmp", "tmp_file", tmp, "original_file", f.destination)
			return nil, fmt.Errorf("recover %s %s: %w", tmp, f.destination, err)
		}
		return nil, err
	}
	if err := fs.Remove(tmp); err != nil {
		f.logger.Warn("failed to remove the tmp", "tmp_file", tmp)
		return nil, fmt.Errorf("remove %s: %w", tmp, err)
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess,
	}
	events = append(events, linkStep)
	return events, nil
}

func (f *FileActionReplace) Type() ActionType {
	return Replace
}

type FileActionBackup struct {
	fileActionBase
	backup string
}

func newFileActionBackup(source, destination, backup string, l *slog.Logger) *FileActionBackup {
	return &FileActionBackup{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
		backup: backup,
	}
}

func (f *FileActionBackup) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Move(f.destination, f.backup); err != nil {
		return nil, err
	}
	var events []ActionEvent
	moveStep := ActionEvent{
		Action:    actionBackup,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.backup),
		EventType: EventStep,
	}
	events = append(events, moveStep)
	if err := fs.Link(f.source, f.destination); err != nil {
		if err := fs.Move(f.backup, f.destination); err != nil {
			f.logger.Warn("failed to restore the backup", "backup_file", f.backup, "original_file", f.destination)
			return nil, fmt.Errorf("recover %s %s: %w", f.backup, f.destination, err)
		}
		return nil, err
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess,
	}
	events = append(events, linkStep)
	return events, nil
}

func (f *FileActionBackup) Type() ActionType {
	return Backup
}

type FileActionAdopt struct {
	fileActionBase
}

func newFileActionAdopt(source, destination string, l *slog.Logger) *FileActionAdopt {
	return &FileActionAdopt{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *FileActionAdopt) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Move(f.destination, f.source); err != nil {
		return nil, err
	}
	var events []ActionEvent
	moveStep := ActionEvent{
		Action:    actionAdopt,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventStep,
	}
	events = append(events, moveStep)
	if err := fs.Link(f.source, f.destination); err != nil {
		if err := fs.Move(f.source, f.destination); err != nil {
			f.logger.Warn("failed to restore the original", "new_file", f.source, "original_file", f.destination)
			return nil, fmt.Errorf("recover %s %s: %w", f.source, f.destination, err)
		}
		return nil, err
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess,
	}
	events = append(events, linkStep)
	return events, nil
}

func (f *FileActionAdopt) Type() ActionType {
	return Adopt
}

type FileActionRemove struct {
	fileActionBase
}

func newFileActionRemove(source, destination string, l *slog.Logger) *FileActionRemove {
	return &FileActionRemove{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *FileActionRemove) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Remove(f.destination); err != nil {
		return nil, err
	}
	return []ActionEvent{
		{
			Action:    actionRemove,
			Msg:       f.destination,
			EventType: EventSuccess,
		},
	}, nil
}

func (f *FileActionRemove) Type() ActionType {
	return Remove
}
