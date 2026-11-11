package apl

// Class is an APL2 error class.
type Class int

const (
	SyntaxError Class = iota + 1
	ValueError
	ValenceError
	DomainError
	LengthError
	RankError
	IndexError
	DefnError
	NonceError
	SystemLimit
	Interrupt
)

var classNames = map[Class]string{
	SyntaxError:  "SYNTAX ERROR",
	ValueError:   "VALUE ERROR",
	ValenceError: "VALENCE ERROR",
	DomainError:  "DOMAIN ERROR",
	LengthError:  "LENGTH ERROR",
	RankError:    "RANK ERROR",
	IndexError:   "INDEX ERROR",
	DefnError:    "DEFN ERROR",
	NonceError:   "NONCE ERROR",
	SystemLimit:  "SYSTEM LIMIT",
	Interrupt:    "INTERRUPT",
}

func (c Class) String() string { return classNames[c] }

// NoPos marks an error without a known source position.
const NoPos = -1

// Error is an APL error.
type Error struct {
	Class Class
	Pos   int    // rune offset in Src, or NoPos
	Src   string // source line, if known
	Where string // function and line number, e.g. "F[2]"
}

func (e *Error) Error() string { return e.Class.String() }

// newError returns an error without position.
func newError(c Class) *Error { return &Error{Class: c, Pos: NoPos} }

// errorAt returns an error at rune offset pos.
func errorAt(c Class, pos int) *Error { return &Error{Class: c, Pos: pos} }
