package profile

import (
	"fmt"
	"testing"
)

// go test -run TestParseLine -v ./profile
func TestParseLine(t *testing.T) {
	line := `git.garena.com/shopee/loan-service/credit_backend/credit-pricing-center/src/biz/pricing_rule_admin_service_biz.go:2839.38,2840.53 1 0`
	block, err := ParseBlock(line)
	if err != nil {
		t.Fatal(err)
	}

	s := fmt.Sprintf("%+v", block)
	exp := `&{FileName:git.garena.com/shopee/loan-service/credit_backend/credit-pricing-center/src/biz/pricing_rule_admin_service_biz.go Block:{Start:{Line:2839 Col:38} End:{Line:2840 Col:53}} NumStmts:1 Count:0}`

	if s != exp {
		t.Fatalf("expect s to be '%s',actual:'%s'", exp, s)
	}
}
