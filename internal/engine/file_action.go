/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
)

const (
	actionLink     = "link"
	actionBackup   = "backup"
	actionSkip     = "skip"
	actionAdopt    = "adopt"
	actionRemove   = "remove"
	actionCreated  = "created"
	actionRestore  = "restore"
	actionLeftover = "leftover"
)

// EventType defines the type of event that a file action has performed
type EventType int

const (
	// EventSuccess is the event type emitted when the operation succeeds
	EventSuccess EventType = iota
	// EventStep is the event type emitted when the operation is a multistep operation and the step is completed
	EventStep
	// EventSkip is the event type emitted when the operation is skipped
	EventSkip
	// EventIgnore is the event type emitted when the operation is ignored
	EventIgnore
	// EventUndo is the event type emitted when the operation is undone
	EventUndo
	// EventFailure is the event type emitted when the operation is at a failed state
	EventFailure
)

// ActionKind defines the kind of file action that needs to be performed
type ActionKind int

const (
	// ActionUpToDate when the action is up-to-date (no operations needed)
	ActionUpToDate ActionKind = iota
	// ActionSkip when the action is to skip the operation
	ActionSkip
	// ActionLink when the action is to link a source to destination
	ActionLink
	// ActionReplace when the action is to replace an existing file
	ActionReplace
	// ActionBackup when the action is to back up the existing file
	ActionBackup
	// ActionAdopt when the action is to copy the destination file into source before linking
	ActionAdopt
	// ActionRemove when the action is to remove the file
	ActionRemove
	// NOTE: Keep to count the number of actions. Should always the last element
	numActionKinds
)

const corruptedSystemErrMsg = "failed to undo; system is in a corrupted state; manual intervention needed"

// ActionEvent stores the event that occurred while performing an action
type ActionEvent struct {
	Action    string
	Msg       string
	EventType EventType
}

const backupExtension = "bestow.backup"
const tmpExtension = "bestow.tmp"

type fileAction interface {
	execute(fs FileSystem) ([]ActionEvent, error)
	undo(fs FileSystem) ([]ActionEvent, error)
	kind() ActionKind
}

type fileActionPaths struct {
	source      string
	destination string
	logger      *slog.Logger
}

type fileActionUpToDate struct {
	fileActionPaths
	reason string
}

func newFileActionUpToDate(source, destination, reason string, l *slog.Logger) *fileActionUpToDate {
	return &fileActionUpToDate{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
		reason: reason,
	}
}

func (f *fileActionUpToDate) execute(_ FileSystem) ([]ActionEvent, error) {
	f.logger.Debug(f.reason, "source", f.source, "destination", f.destination)
	return []ActionEvent{
		{EventType: EventIgnore},
	}, nil
}

func (f *fileActionUpToDate) undo(_ FileSystem) ([]ActionEvent, error) {
	return []ActionEvent{
		{EventType: EventIgnore},
	}, nil
}

func (f *fileActionUpToDate) kind() ActionKind {
	return ActionUpToDate
}

type fileActionSkip struct {
	fileActionPaths
	reason string
}

func newFileActionSkip(source, destination, reason string, l *slog.Logger) *fileActionSkip {
	return &fileActionSkip{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
		reason: reason,
	}
}

func (f *fileActionSkip) execute(_ FileSystem) ([]ActionEvent, error) {
	return []ActionEvent{
		{
			Action:    actionSkip,
			Msg:       fmt.Sprintf("%s -> %s [%s]", f.source, f.destination, f.reason),
			EventType: EventSkip,
		},
	}, nil
}

func (f *fileActionSkip) undo(_ FileSystem) ([]ActionEvent, error) {
	return []ActionEvent{
		{EventType: EventIgnore},
	}, nil
}

func (f *fileActionSkip) kind() ActionKind {
	return ActionSkip
}

type fileActionLink struct {
	fileActionPaths
}

func newFileActionLink(source, destination string, l *slog.Logger) *fileActionLink {
	return &fileActionLink{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionLink) execute(fs FileSystem) ([]ActionEvent, error) {
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

func (f *fileActionLink) undo(fs FileSystem) ([]ActionEvent, error) {
	if err := fs.Remove(f.destination); err != nil {
		f.logger.Warn("failed to undo; manual intervention needed", "action", "unlink", "path", f.destination)
		return nil, err
	}
	return []ActionEvent{
		{
			Action:    actionRemove,
			Msg:       f.destination,
			EventType: EventUndo,
		},
	}, nil
}

func (f *fileActionLink) kind() ActionKind {
	return ActionLink
}

type fileActionReplace struct {
	fileActionPaths
}

func newFileActionReplace(source, destination string, l *slog.Logger) *fileActionReplace {
	return &fileActionReplace{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionReplace) execute(fs FileSystem) ([]ActionEvent, error) {
	var events []ActionEvent
	tmp := fmt.Sprintf("%s.%s", f.destination, tmpExtension)
	if err := fs.Move(f.destination, tmp); err != nil {
		return nil, err
	}
	moveStep := ActionEvent{
		Action:    actionRemove,
		Msg:       f.destination,
		EventType: EventStep,
	}
	events = append(events, moveStep)
	if err := fs.Link(f.source, f.destination); err != nil {
		if err := fs.Move(tmp, f.destination); err != nil {
			f.logger.Warn("failed to restore the tmp; manual recovery needed", "tmp_file", tmp, "original_file", f.destination)
			return events, fmt.Errorf("recover %s %s: %w", tmp, f.destination, err)
		}
		// Return no event since technically no changes are done after reverting temp
		return nil, err
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess, // This is the success event. The next event is a step of housekeeping
	}
	events = append(events, linkStep)
	if err := fs.Remove(tmp); err != nil {
		f.logger.Warn("failed to remove the temp file", "tmp_file", tmp)
		tmpStep := ActionEvent{
			Action:    actionLeftover,
			Msg:       fmt.Sprintf("temp file %s", tmp),
			EventType: EventFailure,
		}
		events = append(events, tmpStep)
		return events, nil // Returning nil here since the intended action is complete, although the temp file is there.
	}
	removeStep := ActionEvent{
		Action:    actionRemove,
		Msg:       tmp,
		EventType: EventIgnore, // This is a housekeeping event, that should be ignored in summary
	}
	events = append(events, removeStep)
	return events, nil
}

func (f *fileActionReplace) undo(fs FileSystem) ([]ActionEvent, error) {
	f.logger.Warn("undo will not recover the original file", "path", f.destination)
	if err := fs.Remove(f.destination); err != nil {
		f.logger.Warn("failed to undo; manual intervention needed", "action", "remove", "path", f.destination)
		return nil, err
	}
	return []ActionEvent{
		{
			Action:    actionRemove,
			Msg:       f.destination,
			EventType: EventUndo,
		},
	}, nil
}

func (f *fileActionReplace) kind() ActionKind {
	return ActionReplace
}

type fileActionBackup struct {
	fileActionPaths
	backup string
}

func newFileActionBackup(source, destination, backup string, l *slog.Logger) *fileActionBackup {
	return &fileActionBackup{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
		backup: backup,
	}
}

func (f *fileActionBackup) execute(fs FileSystem) ([]ActionEvent, error) {
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
			return events, fmt.Errorf("recover %s %s: %w", f.backup, f.destination, err)
		}
		// No events emitted since technically nothing happened
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

func (f *fileActionBackup) undo(fs FileSystem) ([]ActionEvent, error) {
	if err := fs.Move(f.backup, f.destination); err != nil {
		return nil, err
	}
	return []ActionEvent{
		{
			Action:    actionRestore,
			Msg:       fmt.Sprintf("%s -> %s", f.backup, f.destination),
			EventType: EventUndo,
		},
	}, nil
}

func (f *fileActionBackup) kind() ActionKind {
	return ActionBackup
}

type fileActionAdopt struct {
	fileActionPaths
}

func newFileActionAdopt(source, destination string, l *slog.Logger) *fileActionAdopt {
	return &fileActionAdopt{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionAdopt) execute(fs FileSystem) ([]ActionEvent, error) {
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
			return events, fmt.Errorf("recover %s %s: %w", f.source, f.destination, err)
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

func (f *fileActionAdopt) undo(fs FileSystem) ([]ActionEvent, error) {
	if err := fs.Remove(f.destination); err != nil {
		f.logger.Warn(corruptedSystemErrMsg, "action", "move", "source", f.source, "destination", f.destination)
		return nil, fmt.Errorf("file system in corrupted state; manual intervention needed: %w", err)
	}
	removeStep := ActionEvent{
		Action:    actionRemove,
		Msg:       f.destination,
		EventType: EventUndo,
	}
	if err := fs.Move(f.source, f.destination); err != nil {
		f.logger.Warn(corruptedSystemErrMsg, "action", "move", "source", f.source, "destination", f.destination)
		return []ActionEvent{removeStep}, err
	}
	moveStep := ActionEvent{
		Action:    actionRestore,
		Msg:       fmt.Sprintf("%s -> %s", f.source, f.destination),
		EventType: EventUndo,
	}
	return []ActionEvent{removeStep, moveStep}, nil
}

func (f *fileActionAdopt) kind() ActionKind {
	return ActionAdopt
}

type fileActionRemove struct {
	fileActionPaths
}

func newFileActionRemove(source, destination string, l *slog.Logger) *fileActionRemove {
	return &fileActionRemove{
		fileActionPaths: fileActionPaths{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionRemove) execute(fs FileSystem) ([]ActionEvent, error) {
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

func (f *fileActionRemove) undo(fs FileSystem) ([]ActionEvent, error) {
	if err := fs.Link(f.source, f.destination); err != nil {
		f.logger.Warn(corruptedSystemErrMsg, "action", "link", "source", f.source, "destination", f.destination)
		return nil, err
	}
	return []ActionEvent{
		{
			Action:    actionLink,
			Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
			EventType: EventUndo,
		},
	}, nil
}

func (f *fileActionRemove) kind() ActionKind {
	return ActionRemove
}
