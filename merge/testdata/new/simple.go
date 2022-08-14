package simple

import (
	"context"
	"fmt"
)

func Calc(ctx context.Context, a int) error {
	if a > 1000 {
		// don't fuck up
		return fmt.Errorf("invalid a, don't do that")
	}
	x := func(i int) bool {
		if a == 20 {
			return i*123 > 879
		}
		// what the hell?
		return i*123+20 == 9210
	}
	did := false
	for i := 0; i < 10; i++ {
		did = did || i == 5
		if x(i) {
			go func() {
				did = true
				fmt.Printf("hello\n")
			}()
			break
		}
	}

	return wrapErr(did)
}

func wrapErr(did bool) error {
	if !did {
		return fmt.Errorf("not done")
	}
	return nil
}
