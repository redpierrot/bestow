/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	actionLink    = "link"
	actionBackup  = "backup"
	actionSkip    = "skip"
	actionAdopt   = "adopt"
	actionRemove  = "remove"
	actionCreated = "created"
	actionRestore = "restore"
)

type EventType int

const (
	EventSuccess EventType = iota
	EventStep
	EventSkip
	EventWarn
	EventIgnore
	EventUndo
)

type ActionKind int

const (
	UpToDate ActionKind = iota
	Skip
	Link
	Replace
	Backup
	Adopt
	Remove
)

const corruptedSystemErrMsg = "failed to undo; system is in a corrupted state; manual intervention needed"

func (a ActionKind) String() string {
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

const backupExtension = "bestow.backup"
const tmpExtension = "bestow.tmp"

type fileAction interface {
	preflight(fs FileSystem) error
	execute(fs FileSystem) ([]ActionEvent, error)
	undo(fs FileSystem) ([]ActionEvent, error)
	kind() ActionKind
}

type fileActionBase struct {
	source      string
	destination string
	logger      *slog.Logger
}

type fileActionUpToDate struct {
	fileActionBase
	reason string
}

func newFileActionUpToDate(source, destination, reason string, l *slog.Logger) *fileActionUpToDate {
	return &fileActionUpToDate{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
		reason: reason,
	}
}

func (f *fileActionUpToDate) preflight(_ FileSystem) error {
	return nil
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
	return UpToDate
}

type fileActionSkip struct {
	fileActionBase
	reason string
}

func newFileActionSkip(source, destination, reason string, l *slog.Logger) *fileActionSkip {
	return &fileActionSkip{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
		reason: reason,
	}
}

func (f *fileActionSkip) preflight(_ FileSystem) error {
	return nil
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
	return Skip
}

type fileActionLink struct {
	fileActionBase
}

func newFileActionLink(source, destination string, l *slog.Logger) *fileActionLink {
	return &fileActionLink{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionLink) preflight(fs FileSystem) error {
	if err := fs.Readable(f.source); err != nil {
		return err
	}
	existingDestPath, err := existingParent(f.destination, fs)
	if err != nil {
		return err
	}
	if err := fs.Writable(existingDestPath); err != nil {
		return err
	}
	return nil
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
	return Link
}

type fileActionReplace struct {
	fileActionBase
}

func newFileActionReplace(source, destination string, l *slog.Logger) *fileActionReplace {
	return &fileActionReplace{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionReplace) preflight(fs FileSystem) error {
	if err := fs.Readable(f.source); err != nil {
		return err
	}
	destParent := filepath.Dir(f.destination)
	if err := fs.Writable(destParent); err != nil {
		return err
	}
	return nil
}

func (f *fileActionReplace) execute(fs FileSystem) ([]ActionEvent, error) {
	var events []ActionEvent
	tmp := fmt.Sprintf("%s.%s", f.destination, tmpExtension)
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
			f.logger.Warn("failed to restore the tmp; manual recovery needed", "tmp_file", tmp, "original_file", f.destination)
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
	return Replace
}

type fileActionBackup struct {
	fileActionBase
	backup string
}

func newFileActionBackup(source, destination, backup string, l *slog.Logger) *fileActionBackup {
	return &fileActionBackup{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
		backup: backup,
	}
}

func (f *fileActionBackup) preflight(fs FileSystem) error {
	if err := fs.Readable(f.source); err != nil {
		return err
	}
	destParent := filepath.Dir(f.destination)
	if err := fs.Writable(destParent); err != nil {
		return err
	}
	return nil
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
	return Backup
}

type fileActionAdopt struct {
	fileActionBase
}

func newFileActionAdopt(source, destination string, l *slog.Logger) *fileActionAdopt {
	return &fileActionAdopt{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionAdopt) preflight(fs FileSystem) error {
	srcParent, err := existingParent(f.source, fs)
	if err != nil {
		return err
	}
	if err := fs.Writable(srcParent); err != nil {
		return err
	}
	destParent := filepath.Dir(f.destination)
	if err := fs.Writable(destParent); err != nil {
		return err
	}
	return nil
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

func (f *fileActionAdopt) undo(fs FileSystem) ([]ActionEvent, error) {
	if err := fs.Remove(f.destination); err != nil {
		// TODO: Return an error with corrupted state message
		f.logger.Warn(corruptedSystemErrMsg, "action", "move", "source", f.source, "destination", f.destination)
		return nil, err
	}
	removeStep := ActionEvent{
		Action:    actionRemove,
		Msg:       f.destination,
		EventType: EventUndo,
	}
	if err := fs.Move(f.source, f.destination); err != nil {
		f.logger.Warn(corruptedSystemErrMsg, "action", "move", "source", f.source, "destination", f.destination)
		return nil, err
	}
	moveStep := ActionEvent{
		Action:    actionRestore,
		Msg:       fmt.Sprintf("%s -> %s", f.source, f.destination),
		EventType: EventUndo,
	}
	return []ActionEvent{removeStep, moveStep}, nil
}

func (f *fileActionAdopt) kind() ActionKind {
	return Adopt
}

type fileActionRemove struct {
	fileActionBase
}

func newFileActionRemove(source, destination string, l *slog.Logger) *fileActionRemove {
	return &fileActionRemove{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
			logger:      l,
		},
	}
}

func (f *fileActionRemove) preflight(fs FileSystem) error {
	destParent := filepath.Dir(f.destination)
	if err := fs.Writable(destParent); err != nil {
		return err
	}
	return nil
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
	return Remove
}

func existingParent(path string, fs FileSystem) (string, error) {
	exists, err := fs.Exists(path)
	if err != nil {
		return "", err
	}
	if !exists {
		parent := filepath.Dir(path)
		if filepath.Clean(parent) == string(os.PathSeparator) {
			return parent, nil
		}
		return existingParent(parent, fs)
	}
	return path, nil
}
