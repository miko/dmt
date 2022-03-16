package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	verbose := viper.GetBool("verbose")
	count := 0
	done := false
	var err error
	var buf []byte
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
			if verbose {
				fmt.Printf("[error] [step=%d] error=%s when POSTing schema\n", count, err.Error())
			}
			continue
		}
		buf, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if verbose {
				fmt.Printf("[error] [step=%d] error=%s when reading response:\n%s\n", count, err.Error(), buf)
			}
			continue
		}
		var r schemaResponse
		err = json.Unmarshal(buf, &r)
		if err != nil {
			if verbose {
				fmt.Printf("[error] [step=%d] error=%s when parsing response:\n%s\n", count, err.Error(), buf)
			}
			continue
		}
		if len(r.Errors) == 0 && r.Data.Code == "Success" {
			done = true
			break
		} else {
			if verbose {
				fmt.Printf("[error] [step=%d] error=%s from server:\n%s\n", count, r.Errors[0].Message)
			}
		}
	}
	if done {
		fmt.Println("Uploaded schema")
		return nil
	} else {
		return errors.New("Cannot upload schema, error: %v" + err)
	}
}

func UploadGraphqlData(filename string) error {
	endpoint := viper.GetString("graphql") + "/graphql"
	header := viper.GetString("header")
	verbose := viper.GetBool("verbose")
	count := 0
	done := false
	var err error
	for {
		count++
		if count > 10 {
			break
		}
		var b []byte
		b, err = GetContent(filename)
		if err != nil {
			return err
		}
		client := graphql.NewClient(endpoint)

		req := graphql.NewRequest(string(b))
		req.Header.Set("Cache-Control", "no-cache")
		if header != "" {
			th := strings.Split(header, "=")
			if verbose {
				fmt.Printf("Using header: %s: %s (%d)\n", th[0], th[1], len(th))
			}
			req.Header.Set(th[0], th[1])
		} else {
			fmt.Println("Skipping header")
		}
		var respData map[string]interface{}
		err = client.Run(context.Background(), req, &respData)
		if err != nil {
			if verbose {
				fmt.Printf("[error] [step=%d] error=%s\n", count, err.Error())
			}
			time.Sleep(time.Second)
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
