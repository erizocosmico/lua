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
import "github.com/milochristiansen/lua/luautil"

//import "runtime"

type parser struct {
	l *lexer
}

// Parse reads Lua source into an AST using the types in this package.
func Parse(source string, line int) (block []Stmt, err error) {
	p := &parser{
		l: newLexer(source, line),
	}

	defer func() {
		if x := recover(); x != nil {
			//fmt.Println("Stack Trace:")
			//buf := make([]byte, 4096)
			//buf = buf[:runtime.Stack(buf, true)]
			//fmt.Printf("%s\n", buf)

			switch e := x.(type) {
			case luautil.Error:
				e.Msg = fmt.Sprintf("%v On Line: %v", e.Msg, p.l.tokenline)
				err = e
			case error:
				err = &luautil.Error{Err: e, Type: luautil.ErrTypWrapped}
			default:
				err = &luautil.Error{Msg: fmt.Sprint(x), Type: luautil.ErrTypEvil}
			}
		}
	}()

	for !p.l.checkLook(tknINVALID) {
		block = append(block, p.statement())
	}
	return block, nil
}

func (p *parser) funcDeclStat(local bool) Stmt {
	p.l.getCurrent(tknFunction)

	// Function declarations are exploded into an explicit assignment statement.
	node := stmtInfo(&Assign{
		LocalFunc: local,
		Targets:   []Expr{nil},
		Values:    []Expr{nil},
	}, p.l.current.Line, p.l.current.Col)

	// Read Name
	var ident Expr
	hasSelf := false
	if local {
		p.l.getCurrent(tknName)
		ident = exprInfo(&ConstIdent{
			Value: p.l.current.Lexeme,
		}, p.l.current.Line, p.l.current.Col)
	} else {
		ident = p.ident()
		if p.l.checkLook(tknColon) {
			hasSelf = true
			p.l.getCurrent(tknColon)
			line := p.l.current.Line
			col := p.l.current.Col
			p.l.getCurrent(tknName)
			ident = exprInfo(&TableAccessor{
				Obj: ident,
				Key: exprInfo(&ConstString{
					Value: p.l.current.Lexeme,
				}, p.l.current.Line, p.l.current.Col),
			}, line, col)
		}
	}
	node.(*Assign).Targets[0] = ident

	// Read Parameters and Block
	node.(*Assign).Values[0] = p.funcDeclBody(hasSelf)
	return node
}

// The block opener must have already been read
func (p *parser) block(enders ...int) []Stmt {
	rtn := []Stmt{}
	for !p.l.checkLook(append(enders, tknINVALID)...) {
		rtn = append(rtn, p.statement())
	}
	p.l.getCurrent(enders...)
	return rtn
}

func (p *parser) statement() Stmt {
	switch p.l.look.Type {
	case tknUnnecessary: // ;
		p.l.getCurrent(tknUnnecessary)
		return stmtInfo(&DoBlock{Block: nil}, p.l.current.Line, p.l.current.Col) // FIXME!
	case tknIf:
		p.l.getCurrent(tknIf)
		line := p.l.current.Line
		col := p.l.current.Col
		node := stmtInfo(&If{
			Cond: p.expression(),
		}, line, col)
		rnode := node
		p.l.getCurrent(tknThen)
		node.(*If).Then = p.block(tknElse, tknElseif, tknEnd)
	loop:
		for {
			switch p.l.current.Type {
			case tknElse:
				node.(*If).Else = p.block(tknEnd)
				break loop
			case tknElseif:
				line := p.l.current.Line
				col := p.l.current.Col
				pnode := node
				node = stmtInfo(&If{
					Cond: p.expression(),
				}, line, col)

				p.l.getCurrent(tknThen)

				node.(*If).Then = p.block(tknElse, tknElseif, tknEnd)

				pnode.(*If).Else = []Stmt{node}
			case tknEnd:
				break loop
			default:
				panic("IMPOSSIBLE")
			}
		}
		return rnode
	case tknComment:
		p.l.getCurrent(tknComment)
		line := p.l.current.Line
		col := p.l.current.Col
		comment := p.l.current.Lexeme
		return stmtInfo(&Comment{Text: comment}, line, col)
	case tknWhile:
		p.l.getCurrent(tknWhile)
		line := p.l.current.Line
		col := p.l.current.Col
		cond := p.expression()
		p.l.getCurrent(tknDo)
		return stmtInfo(&WhileLoop{
			Cond:  cond,
			Block: p.block(tknEnd),
		}, line, col)
	case tknDo:
		p.l.getCurrent(tknDo)
		line := p.l.current.Line
		col := p.l.current.Col
		rtn := p.block(tknEnd)
		return stmtInfo(&DoBlock{Block: rtn}, line, col)
	case tknFor:
		p.l.getCurrent(tknFor)
		line := p.l.current.Line
		col := p.l.current.Col

		// Numeric: var = a, b, c
		counter := ""
		var i, l, s Expr

		// Generic: <vars...> in <expr | expr, expr, expr>
		locals := []string{}
		init := []Expr{}

		p.l.getCurrent(tknName)
		numeric := p.l.checkLook(tknSet)

		if numeric {
			counter = p.l.current.Lexeme
			p.l.getCurrent(tknSet)
			i = p.expression()
			p.l.getCurrent(tknSeperator)
			l = p.expression()
			if p.l.checkLook(tknSeperator) {
				p.l.getCurrent(tknSeperator)
				s = p.expression()
			} else {
				s = exprInfo(&ConstInt{Value: "1"}, p.l.current.Line, p.l.current.Col)
			}
		} else {
			for {
				locals = append(locals, p.l.current.Lexeme)
				if !p.l.checkLook(tknSeperator) {
					break
				}
				p.l.getCurrent(tknSeperator)
				p.l.getCurrent(tknName)
			}
			p.l.getCurrent(tknIn)
			for {
				init = append(init, p.expression())
				if !p.l.checkLook(tknSeperator) {
					break
				}
				p.l.getCurrent(tknSeperator)
			}
		}
		p.l.getCurrent(tknDo)
		if numeric {
			return stmtInfo(&ForLoopNumeric{
				Counter: counter,
				Init:    i,
				Limit:   l,
				Step:    s,
				Block:   p.block(tknEnd),
			}, line, col)
		}
		return stmtInfo(&ForLoopGeneric{
			Locals: locals,
			Init:   init,
			Block:  p.block(tknEnd),
		}, line, col)
	case tknRepeat:
		p.l.getCurrent(tknRepeat)
		line := p.l.current.Line
		col := p.l.current.Col
		blk := p.block(tknUntil)
		return stmtInfo(&RepeatUntilLoop{
			Cond:  p.expression(),
			Block: blk,
		}, line, col)
	case tknFunction:
		return p.funcDeclStat(false)
	case tknLocal:
		p.l.getCurrent(tknLocal)
		line := p.l.current.Line
		col := p.l.current.Col
		if p.l.checkLook(tknFunction) {
			// This is incorrect, "local function f" should translate to "local f; f = function" not "local f = function".
			// The compiler has some special case code to correct this.
			return p.funcDeclStat(true)
		}
		targets := []Expr{}
		c := 0
		for !p.l.checkLook(tknSet) {
			c++
			p.l.getCurrent(tknName)
			targets = append(targets, exprInfo(&ConstIdent{
				Value: p.l.current.Lexeme,
			}, p.l.current.Line, p.l.current.Col))
			if !p.l.checkLook(tknSeperator) {
				break
			}
			p.l.getCurrent(tknSeperator)
		}

		vals := []Expr{}
		if p.l.checkLook(tknSet) {
			p.l.getCurrent(tknSet)
			vals = append(vals, p.expression())
			for p.l.checkLook(tknSeperator) {
				p.l.getCurrent(tknSeperator)
				vals = append(vals, p.expression())
			}
		}
		return stmtInfo(&Assign{
			LocalDecl: true,
			Targets:   targets,
			Values:    vals,
		}, line, col)
	case tknDblColon:
		p.l.getCurrent(tknDblColon)
		line := p.l.current.Line
		col := p.l.current.Col
		p.l.getCurrent(tknName)
		lbl := p.l.current.Lexeme
		p.l.getCurrent(tknDblColon)
		return stmtInfo(&Label{Label: lbl}, line, col)
	case tknReturn:
		p.l.getCurrent(tknReturn)
		line := p.l.current.Line
		col := p.l.current.Col
		items := []Expr{}
		for !p.l.checkLook(tknEnd, tknElse, tknElseif, tknUntil, tknUnnecessary, tknINVALID) {
			items = append(items, p.expression())
			if !p.l.checkLook(tknSeperator) {
				break
			}
			p.l.getCurrent(tknSeperator)
		}
		return stmtInfo(&Return{Items: items}, line, col)
	case tknBreak:
		p.l.getCurrent(tknBreak)
		return stmtInfo(&Goto{Label: "break", IsBreak: true}, p.l.current.Line, p.l.current.Col)
	case tknContinue:
		p.l.getCurrent(tknContinue)
		return stmtInfo(&Goto{Label: "continue", IsContinue: true}, p.l.current.Line, p.l.current.Col)
	case tknGoto:
		p.l.getCurrent(tknGoto)
		line := p.l.current.Line
		col := p.l.current.Col
		p.l.getCurrent(tknName)
		return stmtInfo(&Goto{Label: p.l.current.Lexeme}, line, col)
	case tknOParen:
		p.l.getCurrent(tknOParen)
		ident := p.expression()
		p.l.getCurrent(tknCParen)
		return Stmt(p.funcCall(ident).(*FuncCall))
	default:
		ident := p.suffixedValue()
		line := p.l.current.Line
		col := p.l.current.Col
		if v, ok := ident.(*FuncCall); ok {
			return Stmt(v)
		}

		targets := []Expr{ident}
		for p.l.checkLook(tknSeperator) {
			p.l.getCurrent(tknSeperator)
			targets = append(targets, p.suffixedValue())
		}
		p.l.getCurrent(tknSet)
		vals := []Expr{p.expression()}
		for p.l.checkLook(tknSeperator) {
			p.l.getCurrent(tknSeperator)
			vals = append(vals, p.expression())
		}
		return stmtInfo(&Assign{
			Targets: targets,
			Values:  vals,
		}, line, col)
	}
	panic("UNREACHABLE")
}
