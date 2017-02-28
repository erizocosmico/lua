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

// Assign represents an assignment statement.
type Assign struct {
	stmtBase

	// Is this a local variable declaration statement?
	LocalDecl bool `json:"local_decl"`

	// Special case handling for "local function f() end", this should be treated like "local f; f = function() end".
	LocalFunc bool `json:"local_func"`

	Targets []Expr `json:"targets"`
	Values  []Expr `json:"values"` // If len == 0 no values were given, if len == 1 then the value may be a multi-return function call.
}

// FuncCall is declared in the expression parts file (it is both an Expr and a Stmt).

// DoBlock represents a do block (do ... end).
type DoBlock struct {
	stmtBase

	Block []Stmt `json:"block"`
}

// If represents an if statement.
// 'elseif' statements are encoded as nested if statements.
type If struct {
	stmtBase

	Cond Expr   `json:"cond"`
	Then []Stmt `json:"then"`
	Else []Stmt `json:"else"`
}

// WhileLoop represents a while loop.
type WhileLoop struct {
	stmtBase

	Cond  Expr   `json:"cond"`
	Block []Stmt `json:"block"`
}

// RepeatUntilLoop represents a repeat-until loop.
type RepeatUntilLoop struct {
	stmtBase

	Cond  Expr   `json:"cond"`
	Block []Stmt `json:"block"`
}

// ForLoopNumeric represents a numeric for loop.
type ForLoopNumeric struct {
	stmtBase

	Counter string `json:"counter"`

	Init  Expr `json:"init"`
	Limit Expr `json:"limit"`
	Step  Expr `json:"step"`

	Block []Stmt `json:"block"`
}

// ForLoopGeneric represents a generic for loop.
type ForLoopGeneric struct {
	stmtBase

	Locals []string `json:"locals"`
	Init   []Expr   `json:"expr"` // This will always be adjusted to three return results, but AFAIK there is no actual limit on expression count.

	Block []Stmt `json:"block"`
}

type Goto struct {
	stmtBase

	// True if this Goto is actually a break statement. There is no matching label.
	// If Label is not "break" then this is actually a continue statement (a custom
	// extension that the default lexer/parser does not use).
	IsBreak    bool   `json:"is_break"`
	IsContinue bool   `json:"is_continue"`
	Label      string `json:"label"`
}

type Label struct {
	stmtBase

	Label string `json:"label"`
}

type Return struct {
	stmtBase

	Items []Expr `json:"items"`
}

type Comment struct {
	stmtBase

	Text string `json:"text"`
}

func (Comment) exprMark() {}
