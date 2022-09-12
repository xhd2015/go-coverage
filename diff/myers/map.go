package myers

// the myers diff is based on line mapping
// but if you do not care about line-wise change, only block change,
// you can group all blocks together to make them act like lines

// ComputeBlockMapping
// a block is represented in a string
// the function computes
// the result is a mapping from each index in new to its
// counterpart in old, -1 if new.
func ComputeBlockMapping(oldBlocks []string, newBlocks []string) map[int]int {
	m := make(map[int]int, len(newBlocks))
	operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) {
		m[newLine] = oldLine
	}, nil)
	return m
}

// TODO: currently not used, maybe the only important thing is finding sames, not updates or deletions.
// deleted of old
// func ComputeMapping(oldBlocks []string, newBlocks []string) (sameNewToOld map[int]int, updatedNewToOld map[int]int, deletedOld []int) {
// 	sameNewToOld = make(map[int]int, len(newBlocks))
// 	operations := operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) {
// 		sameNewToOld[newLine] = oldLine
// 	}, func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int) {

// 	})
// 	for _, op := range operations {
// 		switch op.Kind {
// 		case Delete:
// 			for i := op.I1; i <= op.I2; i++ {
// 				deletedOld = append(deletedOld, i)
// 			}
// 		case Insert:

// 		}
// 	}
// 	return
// }
