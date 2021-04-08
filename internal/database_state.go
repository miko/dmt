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
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
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
	var resp *api.Response
	resp, err = dg.NewTxn().Query(context.Background(), GET_STATE_QUERY)
	if err != nil {
		//		log.Fatal(err)
		return
	}
	var rs struct {
		GetState []struct {
			Version int    `json:"version,omitempty"`
			State   string `json:"state,omitempty"`
		} `json:"getState,omitempty"`
	}
	err = json.Unmarshal(resp.Json, &rs)
	if err != nil {
		//		log.Fatal(err)
		return
	}
	if len(rs.GetState) > 0 {
		var decoded []byte
		decoded, err = base64.StdEncoding.DecodeString(rs.GetState[0].State)
		err = json.Unmarshal(decoded, &ds)
		if err != nil {
			//			log.Println(err)
			return ds, err
		}
		if ds.CurrentVersion != rs.GetState[0].Version {
			err = fmt.Errorf("State version %d does not match %d\n", ds.CurrentVersion, rs.GetState[0].Version)
			return
		}
	} else {
		err = fmt.Errorf("Database not initialized")
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

func InitializeDatabase(index, basedir string) (err error) {
	ds := &DatabaseState{CurrentVersion: 0, IndexLocation: index, BaseDir: basedir, Date: time.Now()}
	str, _ := json.Marshal(ds)
	encoded := base64.StdEncoding.EncodeToString(str)
	body, err := UploadUpsertData(fmt.Sprintf(`upsert {
  query {
    q(func: has(dmt.version)) {
      uid
      ver as version: dmt.version
    }
  }

  mutation @if(eq(len(ver),0)) {
    set {
      <0x1> <dmt.version> "0"^^<xs:int> .
      <0x1> <dmt.state> "%s" .
    }
  }
}`, encoded), "application/rdf")
	var r struct {
		Data struct {
			Queries struct {
				Q []struct {
					Version int `json:"version"`
				} `json:"q"`
			} `json:"queries"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}
	if len(r.Data.Queries.Q) > 0 {
		return fmt.Errorf("Databse already initialized at version %d", r.Data.Queries.Q[0].Version)
	}
	return err
}
func InitializeDatabase2(index, basedir string) (err error) {
	dg, cancel := NewClient()
	defer cancel()
	qry := GET_STATE_QUERY_UPSERT
	var mts []*api.Mutation
	ds := &DatabaseState{CurrentVersion: 0, IndexLocation: index, BaseDir: basedir, Date: time.Now()}
	str, _ := json.Marshal(ds)
	encoded := base64.StdEncoding.EncodeToString(str)

	mts = append(mts,
		&api.Mutation{
			SetJson: []byte(NewDqlMutation().
				add("uid", "0x1").
				add("dmt.version", 0).
				add("dmt.state", encoded).
				serialize()),
			CommitNow: true,
			Cond:      `@if(eq(len(ver),0))`,
		})

	/*
		mu := &api.Mutation{
				CommitNow: true,
				SetJson:
			}
	*/
	req := &api.Request{
		CommitNow: true,
		Query:     qry,
		//RespFormat: api.Request_JSON,
		Mutations: mts,
	}
	txn := dg.NewTxn()
	//	fmt.Printf("CALLING: %#v\n", req.Mutations[0].)
	var r *api.Response
	r, err = txn.Do(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		var rs struct {
			GetState []struct {
				LenState []interface{} `json:"len_state"`
			} `json:"getState"`
		}
		err = json.Unmarshal(r.Json, &rs)
		if err != nil {
			fmt.Printf("ERROR, JSON=%s\n", r.Json)
			return
		}
		if len(rs.GetState) == 1 && len(rs.GetState[0].LenState) == 0 {
			err = fmt.Errorf("Not initialized")
			return
		}
		if len(rs.GetState) == 0 {
			fmt.Println("Initialized database")
		}

	}
	fmt.Printf("R=%s\n", r.Json)
	return
}

func UpVersion(targetVersion int, se StateEntry) (err error) {
	now := time.Now()
	//fmt.Printf("UpV ver=%d now=%s type=%s\n", targetVersion, now, se.Type)
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
		switch se.Type {
		case "schema.graphql":
			err = UploadGraphqlSchema(ds.BaseDir + "/" + se.Filename)
			if err != nil {
				fmt.Print(err)
				return
			}
			break

		case "schema.dql":
			err = UploadDqlSchema(ds.BaseDir + "/" + se.Filename)
			if err != nil {
				fmt.Print(err)
				return err
			}
			break

		case "data.graphql":
			err = UploadGraphqlData(ds.BaseDir + "/" + se.Filename)
			if err != nil {
				fmt.Print(err)
				return err
			}

			break

		case "mutation.rdf":
			_, err = UploadUpsert(ds.BaseDir+"/"+se.Filename, "application/rdf")
			if err != nil {
				fmt.Print(err)
				return err
			}
			break
		case "mutation.json":
			_, err = UploadUpsert(ds.BaseDir+"/"+se.Filename, "application/json")
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

func UploadUpsertData(content, contenttype string) ([]byte, error) {
	url := viper.GetString("graphql") + "/mutate?commitNow=true"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(content)))

	req.Header.Set("Content-Type", contenttype)
	//	fmt.Println("mutation", url, contenttype)
	//	fmt.Println(content)
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
