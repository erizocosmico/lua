/*
Copyright 2016-2017 by Milo Christiansen

This software is provided 'as-is', without any express or implied warranty. In
no event will the authors be held liable for any damages arising from the use of
this software.

Permission is granted to anyone to use this software for any purpose, including
commercial applications, and to alter it and redistribute it freely, subject to
the following restrictions:

1. The origin of this software must not be misrepresented; you must not claim
that you wrote the original software. If you use this software in a product, an
acknowledgment in the product documentation would be appreciated but is not
required.

2. Altered source versions must be plainly marked as such, and must not be
misrepresented as being the original software.

3. This notice may not be removed or altered from any source distribution.
*/

package ast

import "fmt"

// Unexported to make it hard to generate impossible operators.
type opTyp int

// Operator type for use with the Operator Expr Node.
const (
	OpAdd opTyp = iota
	OpSub
	OpMul
	OpMod
	OpPow
	OpDiv
	OpIDiv
	OpBinAND
	OpBinOR
	OpBinXOR
	OpBinShiftL
	OpBinShiftR
	OpUMinus
	OpBinNot
	OpNot
	OpLength
	OpConcat

	OpEqual
	OpNotEqual
	OpLessThan
	OpGreaterThan
	OpLessOrEqual
	OpGreaterOrEqual

	OpAnd
	OpOr
)

var opTypNames = [][]byte{
	[]byte("OpAdd"),
	[]byte("OpSub"),
	[]byte("OpMul"),
	[]byte("OpMod"),
	[]byte("OpPow"),
	[]byte("OpDiv"),
	[]byte("OpIDiv"),
	[]byte("OpBinAND"),
	[]byte("OpBinOR"),
	[]byte("OpBinXOR"),
	[]byte("OpBinShiftL"),
	[]byte("OpBinShiftR"),
	[]byte("OpUMinus"),
	[]byte("OpBinNot"),
	[]byte("OpNot"),
	[]byte("OpLength"),
	[]byte("OpConcat"),
	[]byte("OpEqual"),
	[]byte("OpNotEqual"),
	[]byte("OpLessThan"),
	[]byte("OpGreaterThan"),
	[]byte("OpLessOrEqual"),
	[]byte("OpGreaterOrEqual"),
	[]byte("OpAnd"),
	[]byte("OpOr"),
}

func (o opTyp) MarshalText() ([]byte, error) {
	op := int(o)
	if len(opTypNames) <= op || op < 0 {
		return nil, fmt.Errorf("invalid opTyp with value %d", op)
	}

	return opTypNames[int(o)], nil
}

func (o opTyp) String() string {
	name, err := o.MarshalText()
	if err != nil {
		return "INVALID"
	}
	return string(name)
}

// Operator represents an operator and it's operands.
type Operator struct {
	exprBase

	Op    opTyp `json:"type"`
	Left  Expr  `json:"lhs"` // Nil if operator is unary
	Right Expr  `json:"rhs"`
}

// FuncCall represents a function call.
// This has the unique property of being both a Stmt and an Expr.
type FuncCall struct {
	exprBase

	Receiver Expr   `json:"receiver"` // The call receiver if any (the part before the ':')
	Function Expr   `json:"function"` // The function value itself, if Receiver is provided this is the part *after* the colon, else it is the whole name.
	Args     []Expr `json:"args"`
}

func (s *FuncCall) stmtMark() {}

// FuncDecl represents a function declaration.
type FuncDecl struct {
	exprBase

	Params     []string `json:"params"`
	IsVariadic bool     `json:"is_variadic"`

	Source string `json:"source"`

	Block []Stmt `json:"block"`
}

// TableConstructor represents a table constructor.
type TableConstructor struct {
	exprBase

	Keys []Expr `json:"keys"` // A nil key for a particular position means that no key was given.
	Vals []Expr `json:"vals"`
}

// TableAccessor represents a table access expression, one of `a.b` or `a[b]`.
type TableAccessor struct {
	exprBase

	Obj Expr `json:"obj"`
	Key Expr `json:"key"`
}

// Parens represents a pair of parenthesis and the expression inside of them.
type Parens struct {
	exprBase

	Inner Expr `json:"inner"`
}

// ConstInt stores an integer constant.
type ConstInt struct {
	exprBase

	Value string `json:"value"`
}

// ConstFloat stores a floating point constant.
type ConstFloat struct {
	exprBase

	Value string `json:"value"`
}

// ConstString stores a string constant.
type ConstString struct {
	exprBase

	Value string `json:"value"`
}

// ConstIdent stores an identifier constant.
type ConstIdent struct {
	exprBase

	Value string `json:"value"`
}

// ConstBool represents a boolean constant.
type ConstBool struct {
	exprBase

	Value bool `json:"value"`
}

// ConstNil represents the constant "nil".
type ConstNil struct {
	exprBase
}

// ConstVariadic represents the variadic expression element (...).
type ConstVariadic struct {
	exprBase
}
