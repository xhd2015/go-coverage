# Introduction

This library handles go-coverage files.
It provides semantic coverage profile merging function.
Check [merge/merge.go](merge/merge.go) for more details.

Example:

```bash
go test -run TestMergeProfile -v ./merge
```

Old Coverage:
![old](./img/go-coverage-old.jpg)

New Coverage:
![new](./img/go-coverage-new.jpg)

Merged Coverage:
![merged](./img/go-coverage-merged.jpg)

# Algorithm

The algorithm implementation is at [merge/merge.go](merge/merge.go).

First, traverse the old and new AST tree to get each basic block's [cleaned code](./code/clean.go). Simply put, cleaned code is code that gets compiled into assembly, excluding any space and comments.

Then, use [myers diff](./diff/myers/diff.go) to find unchanged blocks, see [ComputeBlockMapping](./diff/myers/map.go).In this step we map all blocks in new AST to their unchanged counterpart in old AST.

Finally, use the new-to-old mapping to merge counters, for unchanged blocks, counters are added together.

The algorithm effectively provides incrimental testing coverage across muiltpe changes(e.g. multiple git commits).

# Diff

```bash
go generate ./...
```

# Required go version: 1.16

When build with go1.14:

```bash
$ with-go1.14 go build ./...
# github.com/dop251/goja/parser
../../gopath/pkg/mod/github.com/dop251/goja@v0.0.0-20221229151140-b95230a9dbad/parser/parser.go:150:9: undefined: os.ReadFile
../../gopath/pkg/mod/github.com/dop251/goja@v0.0.0-20221229151140-b95230a9dbad/parser/statement.go:901:19: undefined: os.ReadFile
note: module requires Go 1.16
```
