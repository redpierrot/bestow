package engine

import "github.com/ThisaruGuruge/bestow/internal/file"

type ConflictResolver interface {
	Resolve(src, dest string, existing file.ExistingType) (ResolveStrategy, error)
}

type StaticResolver struct {
	strategy ResolveStrategy
}

// TODO: Make sure to have a prune method to clear the history
func (sr StaticResolver) Resolve(src, dest string, existing file.ExistingType) (ResolveStrategy, error) {
	return sr.strategy, nil
}
