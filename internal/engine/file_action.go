/*
All Rights Reversed (ɔ)
*/

package engine

import (
	"fmt"

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

func newFileActionUpToDate(source, destination, reason string) *FileActionUpToDate {
	return &FileActionUpToDate{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
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

func newFileActionSkip(source, destination, reason string) *FileActionSkip {
	return &FileActionSkip{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
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

func newFileActionLink(source, destination string) *FileActionLink {
	return &FileActionLink{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
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

func newFileActionReplace(source, destination string) *FileActionReplace {
	return &FileActionReplace{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
		},
	}
}

func (f *FileActionReplace) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Remove(f.destination); err != nil {
		return nil, err
	}
	removeStep := ActionEvent{
		Action:    actionRemove,
		Msg:       f.destination,
		EventType: EventStep,
	}
	if err := fs.Link(f.source, f.destination); err != nil {
		return nil, err
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess,
	}
	return []ActionEvent{removeStep, linkStep}, nil
}

func (f *FileActionReplace) Type() ActionType {
	return Replace
}

type FileActionBackup struct {
	fileActionBase
	backup string
}

func newFileActionBackup(source, destination, backup string) *FileActionBackup {
	return &FileActionBackup{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
		},
		backup: backup,
	}
}

func (f *FileActionBackup) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Move(f.destination, f.backup); err != nil {
		return nil, err
	}
	moveStep := ActionEvent{
		Action:    actionBackup,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.backup),
		EventType: EventStep,
	}
	if err := fs.Link(f.source, f.destination); err != nil {
		return nil, err
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess,
	}
	return []ActionEvent{moveStep, linkStep}, nil
}

func (f *FileActionBackup) Type() ActionType {
	return Backup
}

type FileActionAdopt struct {
	fileActionBase
}

func newFileActionAdopt(source, destination string) *FileActionAdopt {
	return &FileActionAdopt{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
		},
	}
}

func (f *FileActionAdopt) Execute(fs file.System) ([]ActionEvent, error) {
	if err := fs.Move(f.destination, f.source); err != nil {
		return nil, err
	}
	moveStep := ActionEvent{
		Action:    actionAdopt,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventStep,
	}
	if err := fs.Link(f.source, f.destination); err != nil {
		return nil, err
	}
	linkStep := ActionEvent{
		Action:    actionLink,
		Msg:       fmt.Sprintf("%s -> %s", f.destination, f.source),
		EventType: EventSuccess,
	}
	return []ActionEvent{moveStep, linkStep}, nil
}

func (f *FileActionAdopt) Type() ActionType {
	return Adopt
}

type FileActionRemove struct {
	fileActionBase
}

func newFileActionRemove(source, destination string) *FileActionRemove {
	return &FileActionRemove{
		fileActionBase: fileActionBase{
			source:      source,
			destination: destination,
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
