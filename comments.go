package screenjournal

import "strconv"

type (
	CommentID uint64
	Comment   string
)

func (id CommentID) UInt64() uint64 {
	return uint64(id)
}

func (id CommentID) String() string {
	return strconv.FormatUint(id.UInt64(), 10)
}

func (c Comment) String() string {
	return string(c)
}
