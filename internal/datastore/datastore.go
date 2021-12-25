package datastore

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
)

type FindAllTodoFilter struct {
	Cursor   string
	Backward bool
	Size     uint64
}

const MaxSize = 25

func (f *FindAllTodoFilter) GetSize() uint64 {
	if f.Size == 0 || f.Size > MaxSize {
		f.Size = MaxSize
	}

	return f.Size
}
