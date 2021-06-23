package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	//"path/filepath"
	"time"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type CancelFunc func()

var cnt int

const (
	GET_STATE_QUERY = `{getState (func:has(dmt.version)) {
  version: dmt.version
  state: dmt.state
 }
}`
	GET_STATE_QUERY_UPSERT = `{getState (func:has(dmt.version)) {
  ver as version: dmt.version
  state: dmt.state
 }
  len_state (func:uid(ver)) {
    count(ver)
  }

}`
	GET_STATE_QUERY_UPSERT_FILTERED = `{getState (func:has(dmt.version))  @filter(eq(dmt.version,%d)) {
  ver as version: dmt.version
  state: dmt.state
 }
  len_state (func:uid(ver)) {
    count(ver)
  }
}`
)

func NewClient() (*dgo.Dgraph, CancelFunc) {
	addr := viper.GetString("dgraph")
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	verbose := viper.GetBool("verbose")
	if verbose {
		fmt.Printf("[info] Connecting to dgraph %s\n", addr)
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	dc := api.NewDgraphClient(conn)
	dg := dgo.NewDgraphClient(dc)

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}
	}

}

func GetDatabaseState() (ds DatabaseState, err error) {
	dg, cancel := NewClient()
	defer cancel()
	return GetDatabaseStateForClient(dg)
}

func GetDatabaseStateForClient(dg *dgo.Dgraph) (ds DatabaseState, err error) {
	verbose := viper.GetBool("verbose")
	if verbose {
		defer func() {
			if err != nil {
				fmt.Printf("[debug] Database version: %d error=%s\n", ds.CurrentVersion, err.Error())
			}
		}()
	}
	var resp *api.Response
	resp, err = dg.NewTxn().Query(context.Background(), GET_STATE_QUERY)
	if err != nil {
		//		log.Fatal(err)
		return
	}
	//fmt.Printf("[debug] JSON response: %s\n", string(resp.Json))
	var rs struct {
		GetState []struct {
			Version int    `json:"version,omitempty"`
			State   string `json:"state,omitempty"`
		} `json:"getState,omitempty"`
	}
	err = json.Unmarshal(resp.Json, &rs)
	if err != nil {
		return
	}
	if len(rs.GetState) > 0 {
		var decoded []byte
		decoded, err = base64.StdEncoding.DecodeString(rs.GetState[0].State)
		err = json.Unmarshal(decoded, &ds)
		if err != nil {
			return ds, err
		}
		if ds.CurrentVersion != rs.GetState[0].Version {
			err = fmt.Errorf("State version %d does not match %d\n", ds.CurrentVersion, rs.GetState[0].Version)
			return
		}
	} else {
		ds.CurrentVersion = -1
		return
	}
	return
}

func DropAll() error {
	dg, cancel := NewClient()
	defer cancel()
	op := api.Operation{DropAll: true}
	ctx := context.Background()
	if err := dg.Alter(ctx, &op); err != nil {
		log.Fatal(err)
	}
	return nil
}

func InitializeDatabase(index string) (err error) {
	verbose := viper.GetBool("verbose")
	ds := &DatabaseState{CurrentVersion: 0, IndexLocation: index, Date: time.Now()}
	str, _ := json.Marshal(ds)
	encoded := base64.StdEncoding.EncodeToString(str)
	done := false
	type res_type struct {
		Data struct {
			Queries struct {
				HasInitialized []struct {
					Version     int  `json:"version,omitempty"`
					Initialized bool `json:"initialized,omitempty"`
				} `json:"has_initialized"`
			} `json:"queries"`
		} `json:"data"`
	}

	for count := 1; count < 6; count++ {
		time.Sleep(time.Second)
		var res res_type
		body, err := UploadUpsertData(`upsert {
  query {
    has_initialized(func: has(dmt.initialized) ) {
      has_initialized as uid
      initialized: dmt.initialized
      version: dmt.version
    }
  }

  mutation @if(eq(len(has_initialized),0) ) {
    set {
      <0x1> <dmt.initialized> "false"^^<xs:boolean> .
    }
  }
}`, "application/rdf")
		if err != nil {
			if err.Error() == "Uid: [1] cannot be greater than lease: [0]" {
				if verbose {
					fmt.Printf("[warn] [step %d] DB not initialized - creating first record\n", count)
				}
				body, err = UploadUpsertData(`upsert {
  query {
    has_initialized(func: has(dmt.initialized) ) {
      has_initialized as uid
      initialized: dmt.initialized
      version: dmt.version
    }
  }

  mutation @if(eq(len(has_initialized),0) ) {
    set {
      <_:blank> <dmt.initialized> "false"^^<xs:boolean> .
    }
  }
}`, "application/rdf")
				if err != nil {
					if verbose {
						fmt.Printf("[error] [step %d] Got error: %s\n", count, err.Error())
					}
					continue
				}
			} else {
				if verbose {
					fmt.Printf("[error] [step %d] Got error: %s\n", count, err.Error())
				}
				continue
			}
		}
		err = json.Unmarshal(body, &res)
		if err != nil {
			if verbose {
				fmt.Printf("[error] [step %d] Got error: %s\n", count, err.Error())
			}
			continue
		}
		if len(res.Data.Queries.HasInitialized) == 0 {
			if verbose {
				fmt.Printf("[warn] [step %d] Got no database initialization info\n", count)
			}
			continue
		} else {
			if verbose {
				fmt.Printf("[info] [step %d] Got initialization data\n", count)
			}
			if res.Data.Queries.HasInitialized[0].Initialized == true {
				if verbose {
					fmt.Printf("[info] [step %d] Database already initialized at version %d\n", count, res.Data.Queries.HasInitialized[0].Version)
				}
				return fmt.Errorf("Database already initialized at version %d", res.Data.Queries.HasInitialized[0].Version)
			}
			done = true
			break
		}
	}
	if !done {
		return fmt.Errorf("Cannot initialize database")
	}
	body, err := UploadUpsertData(fmt.Sprintf(`upsert {
  query {
    q(func: has(dmt.version)) {
      uid
      ver as version: dmt.version
      initialized: dmt.initialized
    }
  }

  mutation @if( eq(len(ver),0) ) {
    set {
      <0x1> <dmt.initialized> "true"^^<xs:boolean> .
      <0x1> <dmt.version> "0"^^<xs:int> .
      <0x1> <dmt.state> "%s" .
    }
  }
}`, encoded), "application/rdf")

	var res res_type
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	return err
}

func UpVersion(targetVersion int, se StateEntry) (err error) {
	now := time.Now()
	verbose := viper.GetBool("verbose")
	if verbose {
		fmt.Printf("[info] Upgrading version to %d - type %s\n", targetVersion, se.Type)
	}
	se.Date = &now
	ds, err := GetDatabaseState()
	if err != nil {
		fmt.Println(err)
		return err
	}
	ds.Entries = append(ds.Entries, se)
	ds.CurrentVersion = targetVersion

	dg, cancel := NewClient()
	defer cancel()

	qry := fmt.Sprintf(GET_STATE_QUERY_UPSERT_FILTERED, targetVersion-1)
	var mts []*api.Mutation

	str, _ := json.Marshal(ds)
	encoded := base64.StdEncoding.EncodeToString(str)

	mts = append(mts,
		&api.Mutation{
			SetJson: []byte(NewDqlMutation().
				add("uid", "0x1").
				add("dmt.version", targetVersion).
				add("dmt.state", encoded).
				serialize()),
			CommitNow: false,
			Cond:      `@if(eq(len(ver),1))`,
		})

	req := &api.Request{
		CommitNow: false,
		Query:     qry,
		Mutations: mts,
	}
	txn := dg.NewTxn()
	defer txn.Discard(context.Background())

	_, err = txn.Do(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		dir, _ := SplitIndex(ds.IndexLocation)
		if fdir := viper.GetString("index"); fdir != "" {
			dir, _ = SplitIndex(fdir)
		}
		if verbose {
			fmt.Printf("[info] dir set to %s\n", dir)
		}
		switch se.Type {
		case "schema.graphql":
			err = UploadGraphqlSchema(dir + "/" + se.Filename)
			if err != nil {
				fmt.Print(err)
				return
			}
			break

		case "schema.dql":
			err = UploadDqlSchema(dir + "/" + se.Filename)
			if err != nil {
				fmt.Print(err)
				return err
			}
			break

		case "data.graphql":
			err = UploadGraphqlData(dir + "/" + se.Filename)
			if err != nil {
				fmt.Print(err)
				return err
			}

			break

		case "mutation.rdf":
			_, err = UploadUpsert(dir+"/"+se.Filename, "application/rdf")
			if err != nil {
				fmt.Print(err)
				return err
			}
			break
		case "mutation.json":
			_, err = UploadUpsert(dir+"/"+se.Filename, "application/json")
			if err != nil {
				fmt.Print(err)
				return err
			}
			break

		default:
			fmt.Printf("Skipping file %s of type %s\n", se.Filename, se.Type)
			return
		}
		err = txn.Commit(context.Background())
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Updated database to version %d\n", targetVersion)
	}
	return
}

func UploadUpsert(filename, contenttype string) ([]byte, error) {
	data, err := GetContent(filename)
	if err != nil {
		return nil, err
	}
	return UploadUpsertData(string(data), contenttype)
}

func ExportData() ([]byte, error) {
	url := viper.GetString("graphql") + "/admin?commitNow=true"
	export_url := viper.GetString("export_url")
	export_access_key := viper.GetString("export_access_key")
	export_secret_key := viper.GetString("export_secret_key")
	if export_url == "" {
		return nil, fmt.Errorf("No export URL")
	}
	if export_access_key == "" {
		return nil, fmt.Errorf("No export_access_key")
	}
	if export_secret_key == "" {
		return nil, fmt.Errorf("No export_secret_key")
	}
	verbose := viper.GetBool("verbose")

	var jsonStr = fmt.Sprintf(`
mutation {
  export(input: {
    destination: "%s"
    accessKey: "%s"
    secretKey: "%s"
  }) {
    response {
      message
      code
    }
  }
}
`, export_url, export_access_key, export_secret_key)
	if verbose {
		fmt.Printf("[info] Exporting dgraph data to URL %s\n", export_url)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("Content-Type", "application/graphql")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		panic(err)
	}
	defer resp.Body.Close()

	if verbose {
		fmt.Println("[info] Server response status:", resp.Status)
	}
	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("Server returned status %d", resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if verbose {
		fmt.Printf("[info] Got server answer:\n%s\n", body)
	}
	var r struct {
		Data struct {
			Export struct {
				Response struct {
					Message string `json:"message,omitempty"`
					Code    string `json:"code,omitempty"`
				} `json:"response,omitempty"`
			} `json:"export,omitempty"`
		} `json:"data,omitempty"`
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	if r.Data.Export.Response.Code != "Success" {
		return nil, fmt.Errorf("%s", r.Data.Export.Response.Message)
	}
	return body, nil
}

func UploadUpsertData(content, contenttype string) ([]byte, error) {
	url := viper.GetString("graphql") + "/mutate?commitNow=true"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(content)))

	req.Header.Set("Content-Type", contenttype)
	verbose := viper.GetBool("verbose")
	if verbose {
		fmt.Printf("[info] Mutating to URL %s - type %s data:\n%s\n", url, contenttype, content)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		return nil, fmt.Errorf("Server returned status %d", resp.Status)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if verbose {
		fmt.Printf("[info] Got server answer:\n%s\n", body)
	}
	var r struct {
		Errors []struct {
			Message string `json:"message,omitempty"`
		} `json:"errors,omitempty"`
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	if len(r.Errors) > 0 {
		return nil, fmt.Errorf("%s", r.Errors[0].Message)
	}
	return body, nil
}

func UploadDqlSchema(filename string) error {
	url := viper.GetString("graphql") + "/alter"
	data, err := GetContent(filename)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var r struct {
		Errors []struct {
			Message string `json:"message,omitempty"`
		} `json:"errors,omitempty"`
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}
	if len(r.Errors) > 0 {
		return fmt.Errorf("%s", r.Errors[0].Message)
	}

	return nil
}
