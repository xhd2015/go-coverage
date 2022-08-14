package simple

import (
	"fmt"
)

func Calc(a int, loop bool) error {
	if a > 10 {
		return fmt.Errorf("invalid a, don't do that")
	}
	did := false
	if loop {
		for i := 0; i < 10; i++ {
			if func(i int) bool {
				if a == 20 {
					return i*123 > 879
				}
				return i*123+20 == 9210
			}(i) {
				go func() {
					did = true
					fmt.Printf("hello\n")
				}()
				break
			}
			did = i > 20
		}
	}
	if !did {
		return fmt.Errorf("not done")
	}
	return nil
}
