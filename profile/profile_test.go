package profile

import (
	"encoding/json"
	"testing"
)

// go test -run TestProfileParse -v ./profile
func TestProfileParse(t *testing.T) {
	profile, err := ParseProfileFile("./testdata/cov.out")
	if err != nil {
		t.Fatal(err)
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}

	exp := `{"Mode":"set","Blocks":[{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":11,"Col":48},"End":{"Line":11,"Col":381},"NumStmts":4,"Count":0},{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":11,"Col":427},"End":{"Line":12,"Col":14},"NumStmts":1,"Count":0},{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":15,"Col":2},"End":{"Line":15,"Col":23},"NumStmts":1,"Count":0},{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":12,"Col":14},"End":{"Line":14,"Col":3},"NumStmts":1,"Count":0},{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":16,"Col":13},"End":{"Line":17,"Col":11},"NumStmts":1,"Count":0},{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":18,"Col":10},"End":{"Line":19,"Col":36},"NumStmts":1,"Count":0},{"FileName":"github.com/example/cov/utils/err_util.go","Start":{"Line":23,"Col":29},"End":{"Line":23,"Col":345},"NumStmts":3,"Count":0}]}`
	if string(profileJSON) != exp {
		t.Fatalf("expect %s = %+v, actual:%+v", `profileJSON`, exp, string(profileJSON))
	}
}
