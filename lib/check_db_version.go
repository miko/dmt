package lib

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/miko/dmt/internal"
)

func CheckDatabaseVersion(dg *dgo.Dgraph) (version int, err error) {
	var ds internal.DatabaseState
	ds, err = internal.GetDatabaseStateForClient(dg)
	if err != nil {
		return
	}
	version = ds.CurrentVersion
	return
}
