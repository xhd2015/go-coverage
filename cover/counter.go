package cover

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Edit interface {
	Insert(pos int, content string)
}

type goCounter struct {
	varName     string
	fset        *token.FileSet
	edit        Edit
	blocks      []Block
	counterStmt func(string) string
}

var _ Callback = ((*goCounter)(nil))

// OnWrapElse implements Callback
func (c *goCounter) OnWrapElse(lbrace int, rbrace int) {
	c.edit.Insert(lbrace, "{")
	c.edit.Insert(rbrace, "}")
}

// OnBlock implements Callback
func (c *goCounter) OnBlock(insertPos token.Pos, pos token.Pos, end token.Pos, numStmts int, basicStmts []ast.Stmt) {
	c.edit.Insert(c.offset(insertPos), c.newCounter(pos, end, numStmts)+";")
}

// newCounter creates a new counter expression of the appropriate form.
func (c *goCounter) newCounter(start, end token.Pos, numStmt int) string {
	stmt := c.counterStmt(fmt.Sprintf("%s.Count[%d]", c.varName, len(c.blocks)))
	c.blocks = append(c.blocks, Block{start, end, numStmt})
	return stmt
}

// offset translates a token position into a 0-indexed byte offset.
func (c *goCounter) offset(pos token.Pos) int {
	return c.fset.Position(pos).Offset
}
