package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/seatgeek/graphql"
	"github.com/spf13/viper"
)

type schemaResponse struct {
	Errors []struct {
		Message string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
	Data struct {
		Message string `json:"message,omitempty"`
		Code    string `json:"code,omitempty"`
	} `json:"data,omitempty"`
}

func dropData(endpoint string) error {
	count := 0
	done := false
	for {
		count++
		time.Sleep(time.Second)
		resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer([]byte(`{"drop_all": true}`)))
		if err != nil {
			continue
		} else {
			if resp.StatusCode == 200 {
				done = true
				break
			} else {
				return fmt.Errorf("Cannot drop data: %s", err.Error())
			}
		}
	}
	if done {
		return nil
	} else {
		return errors.New("Cannot drop data")
	}
}

func UploadGraphqlSchema(url string) error {
	endpoint := viper.GetString("graphql") + "/admin/schema"
	count := 0
	done := false
	var err error
	for {
		count++
		if count > 30 {
			break
		}
		time.Sleep(time.Second)
		f, err := OpenContent(url)
		if err != nil {
			return err
		}
		var resp *http.Response
		resp, err = http.Post(endpoint, "application/json", f)
		if err != nil {
			continue
		}
		var buf []byte
		buf, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		var r schemaResponse
		err = json.Unmarshal(buf, &r)
		if err != nil {
			continue
		}
		if len(r.Errors) == 0 && r.Data.Code == "Success" {
			done = true
			break
		}
	}
	if done {
		fmt.Println("Uploaded schema")
		return nil
	} else {
		return errors.New("Cannot upload schema, error: " + err.Error())
	}
}

func UploadGraphqlData(filename string) error {
	endpoint := viper.GetString("graphql") + "/graphql"
	count := 0
	done := false
	var err error
	for {
		count++
		if count > 10 {
			break
		}
		time.Sleep(time.Second)
		var b []byte
		b, err = GetContent(filename)
		if err != nil {
			return err
		}
		client := graphql.NewClient(endpoint)

		req := graphql.NewRequest(string(b))
		req.Header.Set("Cache-Control", "no-cache")
		var respData map[string]interface{}
		err = client.Run(context.Background(), req, &respData)
		if err != nil {
			continue
		}

		done = true
		break
	}
	if done {
		fmt.Println("Uploaded data")
		return nil
	} else {
		return errors.New("Cannot upload data, error: " + err.Error())
	}
}
