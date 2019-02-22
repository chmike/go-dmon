package dmon

// MsgWriter encode and write messages
type MsgWriter interface {
	Write(*Msg) (int, error)
}

// MsgReader reads and decode messages.
type MsgReader interface {
	Read(*Msg) (int, error)
}
