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
	indexFile := "dmt.json"
	is.BaseDir = dir
	is.IndexFile = indexFile
	now := time.Now()
	_ = now
	version := 0

	var lastmd5 string

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fname := file.Name()
		ftype := ""

		switch {
		case fname == "dmt.json":
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
				en := &StateEntry{
					Filename: fname,
					Date:     &now,
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
	return ioutil.WriteFile(dir+"/"+indexFile, data, 0644)
}
