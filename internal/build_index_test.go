package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BuildIndex used to stamp time.Now() onto every entry, including files that had
// not changed. On migrator's ~2250-entry index that turned a one-file edit into a
// ~650-line diff, and two branches that both ran `dmt build` conflicted on nearly
// every line of _dmt.json (oliviaplatform/oliviaapp/migrator!1164).
//
// The index is derived: only filesum/chainsum carry meaning, and Date is read
// nowhere except `dmt info`. So an unchanged entry must keep the date it already
// had.

func writeIndexFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("cannot write %s: %v", name, err)
		}
	}
}

func readIndex(t *testing.T, dir string) map[string]StateEntry {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "_dmt.json"))
	if err != nil {
		t.Fatalf("cannot read index: %v", err)
	}
	var entries []StateEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("cannot decode index: %v", err)
	}
	out := map[string]StateEntry{}
	for _, e := range entries {
		out[e.Filename] = e
	}
	return out
}

func TestBuildIndex_KeepsDateOfUnchangedEntries(t *testing.T) {
	dir := t.TempDir()
	writeIndexFixture(t, dir, map[string]string{
		"001.API1.schema.graphql": "type A { a: String }\n",
		"002.data.graphql":        "mutation { a }\n",
		"003.other.graphql":       "mutation { b }\n",
	})

	if err := BuildIndex(dir); err != nil {
		t.Fatalf("first build: %v", err)
	}
	first := readIndex(t, dir)

	// Make the timestamps distinguishable: a rebuild in the same second would
	// hide the bug this test is about.
	time.Sleep(1100 * time.Millisecond)

	// Change the MIDDLE file. Its own sums change, and chainsum cascades to
	// everything after it -- but the first file is untouched.
	writeIndexFixture(t, dir, map[string]string{
		"002.data.graphql": "mutation { a, changed }\n",
	})
	if err := BuildIndex(dir); err != nil {
		t.Fatalf("second build: %v", err)
	}
	second := readIndex(t, dir)

	untouched := "001.API1.schema.graphql"
	if a, b := first[untouched], second[untouched]; a.MD5SUM != b.MD5SUM || a.ChainSum != b.ChainSum {
		t.Fatalf("%s should not have changed sums", untouched)
	} else if !a.Date.Equal(*b.Date) {
		t.Errorf("%s: date rewritten from %s to %s; an unchanged entry must keep its date",
			untouched, a.Date.Format(time.RFC3339Nano), b.Date.Format(time.RFC3339Nano))
	}

	// The edited file and everything downstream of it genuinely changed, so a
	// fresh date is correct there -- that is where the date carries information.
	for _, name := range []string{"002.data.graphql", "003.other.graphql"} {
		if first[name].ChainSum == second[name].ChainSum {
			t.Fatalf("%s: expected the chainsum to change", name)
		}
		if !second[name].Date.After(*first[name].Date) {
			t.Errorf("%s: a genuinely changed entry should get a fresh date", name)
		}
	}
}

// A first build has no previous index to read; every entry is new and dated now.
func TestBuildIndex_FirstBuildDatesEverything(t *testing.T) {
	dir := t.TempDir()
	writeIndexFixture(t, dir, map[string]string{"001.API1.schema.graphql": "type A { a: String }\n"})

	before := time.Now().Add(-time.Second)
	if err := BuildIndex(dir); err != nil {
		t.Fatalf("build: %v", err)
	}
	for name, e := range readIndex(t, dir) {
		if e.Date == nil || !e.Date.After(before) {
			t.Errorf("%s: first build must produce a fresh date, got %v", name, e.Date)
		}
	}
}

// A corrupt or partially-written index must not break the rebuild: the whole
// point of `dmt build` is to regenerate it.
func TestBuildIndex_ToleratesUnreadablePreviousIndex(t *testing.T) {
	dir := t.TempDir()
	writeIndexFixture(t, dir, map[string]string{"001.API1.schema.graphql": "type A { a: String }\n"})
	if err := os.WriteFile(filepath.Join(dir, "_dmt.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := BuildIndex(dir); err != nil {
		t.Fatalf("build over a corrupt index: %v", err)
	}
	if got := len(readIndex(t, dir)); got != 1 {
		t.Errorf("expected 1 entry after rebuild, got %d", got)
	}
}
