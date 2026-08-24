package internal

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

func BuildIndex(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var is IndexState
	// One path, computed once. It used to be built twice -- `indexFile` already
	// carried `dir`, and the write then did `dir + "/" + indexFile` -- so any
	// invocation other than `dmt build .` from inside the directory wrote to a
	// doubled path and silently did nothing.
	indexPath := filepath.Join(dir, "_dmt.json")
	is.IndexFile = indexPath
	now := time.Now()
	version := 0

	// Previous index, read best-effort: an unchanged entry keeps the date it
	// already had. The index is derived -- only filesum/chainsum carry meaning,
	// and Date is read nowhere but `dmt info` -- so re-stamping every entry on
	// every build is pure churn. On a 2250-entry index it turned a one-file edit
	// into a ~650-line diff and made two branches that both ran `build` conflict
	// on nearly every line.
	previous := map[string]StateEntry{}
	if raw, err := ioutil.ReadFile(indexPath); err == nil {
		var oldEntries []StateEntry
		if json.Unmarshal(raw, &oldEntries) == nil {
			for _, e := range oldEntries {
				previous[e.Filename] = e
			}
		}
	}

	var lastmd5 string

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fname := file.Name()
		ftype := ""

		switch {
		case fname == "_dmt.json":
			break
		case len(fname) >= 14 && fname[len(fname)-14:] == "schema.graphql":
			ftype = "schema.graphql"
			break
		case len(fname) >= 10 && fname[len(fname)-10:] == "schema.dql":
			ftype = "schema.dql"
			break
		case filepath.Ext(fname) == ".graphql":
			ftype = "data.graphql"
			break
		case filepath.Ext(fname) == ".rdf":
			ftype = "mutation.rdf"
			break
		case filepath.Ext(fname) == ".json":
			ftype = "mutation.json"
			break
		default:
			fmt.Printf("Unsupported file: %s ext: %s\n", fname, filepath.Ext(fname))
		}
		if ftype != "" {
			var sum string
			if sum, err = GetMD5(dir + "/" + fname); err == nil {
				chainsum := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d-%s-%s", version, lastmd5, sum))))
				date := &now
				if prev, ok := previous[fname]; ok &&
					prev.MD5SUM == sum && prev.ChainSum == chainsum && prev.Date != nil {
					date = prev.Date
				}
				en := &StateEntry{
					Filename: fname,
					Date:     date,
					Type:     ftype,
					ChainSum: chainsum,
					MD5SUM:   sum,
				}
				is.Entries = append(is.Entries, *en)
				lastmd5 = chainsum
			} else {
				return err
			}
			version = version + 1
		}
	}
	data, err := json.MarshalIndent(is.Entries, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(indexPath, data, 0644)
}
