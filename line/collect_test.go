package line

import (
	"encoding/json"
	"os"
	"testing"
)

// go test -run TestProject -v ./line
func TestProject(t *testing.T) {
	dir := os.Getenv("TEST_DIR")
	if dir == "" {
		t.Fatalf("requires dir")
	}

	oldCommit := os.Getenv("OLD_COMMIT")
	if oldCommit == "" {
		t.Fatalf("requires OLD_COMMIT")
	}

	newCommit := os.Getenv("NEW_COMMIT")
	if newCommit == "" {
		t.Fatalf("requires NEW_COMMIT")
	}

	mapping, err := CollectUnchangedLinesMapping(dir, oldCommit, newCommit)
	if err != nil {
		t.Fatal(err)
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		t.Fatal(err)
	}
	// ioutil.WriteFile("line-mapping.json", mappingJSON, 0777)
	t.Logf("%s", string(mappingJSON))
}
