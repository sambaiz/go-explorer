package gomod

import (
	"golang.org/x/mod/modfile"
	"golang.org/x/xerrors"
)

// Parse go.mod
func Parse(goMod []byte) (*modfile.File, error) {
	f, err := modfile.Parse("", []byte(goMod), nil)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse go.mod: %w", err)
	}
	return f, nil
}
