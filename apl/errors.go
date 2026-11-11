// Error messages from the prototype APLfile, indexed 1..100.
package apl

// Messages holds the prototype error texts; index 0 is unused.
// Codes 16 and 18 were unused in the original and are goapl additions.
var Messages = [...]string{
	0:   "",
	1:   "Invalid character follows negative sign",
	2:   "Digit must follow a decimal point",
	3:   "Extraneous chars. follow function header",
	4:   "Invalid character encountered",
	5:   "Function already defined",
	6:   "Illegal name to right of explicit result",
	7:   "Invalid function/argument name",
	8:   "Result of assignment not valid variable",
	9:   "Invalid function right argument name",
	10:  "Invalid expression",
	11:  "Symbol not found",
	12:  "Statement no. to branch to not integer",
	13:  "Dyadic operator not preceded by primary",
	14:  "Invalid expression within parentheses",
	15:  "Mismatched parentheses",
	16:  "Function calls nested too deep",
	17:  "Left arg of dyadic funct. not a primary",
	18:  "End of input",
	19:  "Value not boolean",
	20:  "Attempted division by zero",
	21:  "Argument not a scalar",
	22:  "Argument is negative",
	23:  "Argument is not an integer",
	24:  "Argument is a scalar or empty vector",
	25:  "Not used",
	26:  "Invalid outer product expression",
	27:  "Invalid inner product expression",
	28:  "Not used",
	29:  "Left argument is not a vector",
	30:  "Not used",
	31:  "Not used",
	32:  "Error in function argument",
	33:  "Error in function argument",
	34:  "Invalid index expression",
	35:  "Non-scalar indices",
	36:  "Assigned expression not a scalar",
	37:  "Non-integer indices",
	38:  "Index out of range",
	39:  "Invalid index expression",
	40:  "Not used",
	41:  "Not used",
	42:  "Not used",
	43:  "Not used",
	44:  "Not used",
	45:  "Not used",
	46:  "Not used",
	47:  "Not used",
	48:  "Not used",
	49:  "Not used",
	50:  "Number and base of different sign",
	51:  "Argument is a vector of length one",
	52:  "Args. not compatible for inner product",
	53:  "Argument(s) with rank greater than one",
	54:  "Attempted inverse of zero",
	55:  "Args. incompatible for dyadic operation",
	56:  "Left argument not a vector",
	57:  "Not used",
	58:  "Not used",
	59:  "Not used",
	60:  "Greater than three dimensions",
	61:  "Nil",
	62:  "Re-enter last line",
	63:  "Input",
	64:  "Not used",
	65:  "Not used",
	66:  "Not used",
	67:  "Not used",
	68:  "Not used",
	69:  "Not used",
	70:  "Identifier too long",
	71:  "Input line too long",
	72:  "Invalid reduction operator",
	73:  "Dyadic reduction reference",
	74:  "Monadic reference to dyadic operator",
	75:  "Function defined with no statements",
	76:  "Not used",
	77:  "Not used",
	78:  "Not used",
	79:  "Not used",
	80:  "Not used",
	81:  "Not used",
	82:  "Not used",
	83:  "Not used",
	84:  "Not used",
	85:  "Not used",
	86:  "Not used",
	87:  "Not used",
	88:  "Not used",
	89:  "Not used",
	90:  "Not used",
	91:  "Not used",
	92:  "Not used",
	93:  "Not used",
	94:  "Not used",
	95:  "Not used",
	96:  "Not used",
	97:  "Not used",
	98:  "Not used",
	99:  "Not used",
	100: "Variable not assigned a value",
}

// NoPos marks an error without a known source position.
const NoPos = -1

// Error is an APL error with a prototype message code.
type Error struct {
	Code  int
	Pos   int    // rune offset in Src, or NoPos
	Src   string // source line, if known
	Where string // function and line number, e.g. "F[2]"
}

func (e *Error) Error() string { return Messages[e.Code] }

// newError returns an error without position.
func newError(code int) *Error { return &Error{Code: code, Pos: NoPos} }

// errorAt returns an error at rune offset pos.
func errorAt(code, pos int) *Error { return &Error{Code: code, Pos: pos} }
