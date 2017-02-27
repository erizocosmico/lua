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

// Lots of unexported stuff to prevent generation and insertion of invalid/unexpected Node types.
// If you want to use this with a different Lua version it would probably be better to make a copy
// and add what you need directly instead of trying to inject what you need.

// Node represents an item in the AST.
type Node interface {
	nodeMark()
	GetLine() int
	setLine(int)
	GetKind() NodeKind
	setKind(NodeKind)
}

type NodeKind uint

const (
	ErrNode NodeKind = iota
	AssignNode
	DoBlockNode
	IfNode
	WhileNode
	RepeatNode
	ForNumericNode
	ForGenericNode
	GotoNode
	LabelNode
	ReturnNode
	OpNode
	FuncCallNode
	FuncDeclNode
	TableConstructorNode
	TableAccessorNode
	ParensNode
	ConstIntNode
	ConstFloatNode
	ConstStringNode
	ConstIdentNode
	ConstBoolNode
	ConstNilNode
	ConstVariadicNode
)

var nodeKindNames = [][]byte{
	[]byte("Err"),
	[]byte("Assign"),
	[]byte("DoBlock"),
	[]byte("If"),
	[]byte("While"),
	[]byte("Repeat"),
	[]byte("ForNumeric"),
	[]byte("ForGeneric"),
	[]byte("Goto"),
	[]byte("Label"),
	[]byte("Return"),
	[]byte("Op"),
	[]byte("FuncCall"),
	[]byte("FuncDecl"),
	[]byte("TableConstructor"),
	[]byte("TableAccessor"),
	[]byte("Parens"),
	[]byte("ConstInt"),
	[]byte("ConstFloat"),
	[]byte("ConstString"),
	[]byte("ConstIdent"),
	[]byte("ConstBool"),
	[]byte("ConstNil"),
	[]byte("ConstVariadic"),
}

func (o NodeKind) MarshalText() ([]byte, error) {
	op := uint(o)
	if uint(len(nodeKindNames)) <= op || op < 0 {
		return nodeKindNames[0], nil
	}

	return nodeKindNames[int(o)], nil
}

func (o NodeKind) String() string {
	name, _ := o.MarshalText()
	return string(name)
}

type nodeBase struct {
	Kind NodeKind
	Line int
}

func (nodeBase) nodeMark()             {}
func (n nodeBase) GetLine() int        { return n.Line }
func (n *nodeBase) setLine(l int)      { n.Line = l }
func (n nodeBase) GetKind() NodeKind   { return n.Kind }
func (n *nodeBase) setKind(k NodeKind) { n.Kind = k }

// Stmt represents a statement Node.
type Stmt interface {
	Node
	stmtMark()
}

type stmtBase struct {
	nodeBase
}

func (s *stmtBase) stmtMark() {}

// Expr represents an expression element Node.
type Expr interface {
	Node
	exprMark()
}

type exprBase struct {
	nodeBase
}

func (s *exprBase) exprMark() {}

// insert is a helper for inserting a new statement into a block.
// Invalid values for at cause the statement to be appended to the end.
func insert(b []Stmt, at int, s Stmt) []Stmt {
	if at < 0 || at >= len(b) {
		return append(b, s)
	}

	b = append(b, nil)
	copy(b[at+1:], b[at:])
	b[at] = s
	return b
}

// remove is a helper for removing a statement from a block.
// Invalid values for at will cause b to be returned unchanged.
func remove(b []Stmt, at int) []Stmt {
	if at < 0 || at >= len(b) {
		return b
	}

	if at == len(b)-1 {
		return b[:at]
	}

	return append(b[:at], b[at+1:]...)
}

// stmtInfo attaches line information to a Stmt and returns the Stmt.
func stmtInfo(n Stmt, line int) Stmt {
	var k NodeKind
	switch n.(type) {
	case *Assign:
		k = AssignNode
	case *DoBlock:
		k = DoBlockNode
	case *If:
		k = IfNode
	case *WhileLoop:
		k = WhileNode
	case *RepeatUntilLoop:
		k = RepeatNode
	case *ForLoopGeneric:
		k = ForGenericNode
	case *ForLoopNumeric:
		k = ForNumericNode
	case *Goto:
		k = GotoNode
	case *Label:
		k = LabelNode
	case *Return:
		k = ReturnNode
	}
	n.setKind(k)
	n.setLine(line)
	return n
}

// exprInfo attaches line information to a Expr and returns the Expr.
func exprInfo(n Expr, line int) Expr {
	var k NodeKind
	switch n.(type) {
	case *FuncDecl:
		k = FuncDeclNode
	case *FuncCall:
		k = FuncCallNode
	case *Operator:
		k = OpNode
	case *Parens:
		k = ParensNode
	case *TableConstructor:
		k = TableConstructorNode
	case *TableAccessor:
		k = TableAccessorNode
	case *ConstInt:
		k = ConstIntNode
	case *ConstFloat:
		k = ConstFloatNode
	case *ConstBool:
		k = ConstBoolNode
	case *ConstNil:
		k = ConstNilNode
	case *ConstString:
		k = ConstStringNode
	case *ConstVariadic:
		k = ConstVariadicNode
	case *ConstIdent:
		k = ConstIdentNode
	}
	n.setKind(k)
	n.setLine(line)
	return n
}

// Visitor is used with Walk.
type Visitor interface {
	Visit(n Node) Visitor
}

type basicVisitor func(n Node) Visitor

func (f basicVisitor) Visit(n Node) Visitor { return f(n) }

// NewVisitor takes a simple function and turns it into a basic Visitor, ready to use with Walk.
func NewVisitor(f func(n Node) Visitor) Visitor {
	return basicVisitor(f)
}

// Walk traverses the given AST node and it's children in depth-first order.
// For each node it calls the visitor for that level and then uses the returned visitor for the
// child nodes (if any). If the visitor for a given node returns nil that node's children will
// not be visited. Once all of a node's children are visited the visitor for that level is called
// one final time with nil as its argument.
func Walk(v Visitor, n Node) {
	v = v.Visit(n)
	if v == nil {
		return
	}

	switch nn := n.(type) {
	case *Assign:
		for _, nnn := range nn.Targets {
			Walk(v, nnn)
		}
		for _, nnn := range nn.Values {
			Walk(v, nnn)
		}
	case *DoBlock:
		for _, nnn := range nn.Block {
			Walk(v, nnn)
		}
	case *If:
		Walk(v, nn.Cond)
		for _, nnn := range nn.Then {
			Walk(v, nnn)
		}
		for _, nnn := range nn.Else {
			Walk(v, nnn)
		}
	case *WhileLoop:
		Walk(v, nn.Cond)
		for _, nnn := range nn.Block {
			Walk(v, nnn)
		}
	case *RepeatUntilLoop:
		for _, nnn := range nn.Block {
			Walk(v, nnn)
		}
		Walk(v, nn.Cond)
	case *ForLoopNumeric:
		Walk(v, nn.Init)
		Walk(v, nn.Limit)
		Walk(v, nn.Step)
		for _, nnn := range nn.Block {
			Walk(v, nnn)
		}
	case *ForLoopGeneric:
		for _, nnn := range nn.Init {
			Walk(v, nnn)
		}
		for _, nnn := range nn.Block {
			Walk(v, nnn)
		}
	case *Goto:
	case *Label:
	case *Return:
		for _, nnn := range nn.Items {
			Walk(v, nnn)
		}
	case *Operator:
		Walk(v, nn.Left)
		Walk(v, nn.Right)
	case *FuncCall:
		if nn.Receiver != nil {
			Walk(v, nn.Receiver)
		}
		Walk(v, nn.Function)
		for _, nnn := range nn.Args {
			Walk(v, nnn)
		}
	case *FuncDecl:
		for _, nnn := range nn.Block {
			Walk(v, nnn)
		}
	case *TableConstructor:
		for _, nnn := range nn.Keys {
			Walk(v, nnn)
		}
		for _, nnn := range nn.Vals {
			Walk(v, nnn)
		}
	case *TableAccessor:
		Walk(v, nn.Obj)
		Walk(v, nn.Key)
	case *Parens:
		Walk(v, nn.Inner)
	case *ConstInt:
	case *ConstFloat:
	case *ConstString:
	case *ConstIdent:
	case *ConstBool:
	case *ConstNil:
	case *ConstVariadic:
	default:
		panic("IMPOSSIBLE")
	}
	v.Visit(nil)
}

type inspector func(Node) bool

func (f inspector) Visit(n Node) Visitor {
	if f(n) {
		return f
	}
	return nil
}

// Inspect is exactly like Walk, except f is called for each node only if a call to f
// returns true for that node's parent (f is always called for the root node).
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
