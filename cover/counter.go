package cover

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Edit interface {
	Insert(pos int, content string)
}

type Counter struct {
	VarName     string
	Fset        *token.FileSet
	Edit        Edit
	Blocks      []Block
	CounterStmt func(string) string
}

var _ Callback = ((*Counter)(nil))

// OnWrapElse implements Callback
func (c *Counter) OnWrapElse(lbrace int, rbrace int) {
	c.Edit.Insert(lbrace, "{")
	c.Edit.Insert(rbrace, "}")
}

// OnBlock implements Callback
func (c *Counter) OnBlock(insertPos token.Pos, pos token.Pos, end token.Pos, numStmts int, basicStmts []ast.Stmt) {
	c.Edit.Insert(c.offset(insertPos), c.newCounter(pos, end, numStmts)+";")
}

// newCounter creates a new counter expression of the appropriate form.
func (c *Counter) newCounter(start, end token.Pos, numStmt int) string {
	stmt := c.CounterStmt(fmt.Sprintf("%s.Count[%d]", c.VarName, len(c.Blocks)))
	c.Blocks = append(c.Blocks, Block{start, end, numStmt})
	return stmt
}

// offset translates a token position into a 0-indexed byte offset.
func (c *Counter) offset(pos token.Pos) int {
	return c.Fset.Position(pos).Offset
}
