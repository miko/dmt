package internal

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/utahta/go-openuri"
)

func GetIndexState() (is IndexState, err error) {
	var data []byte
	verbose := viper.GetBool("verbose")
	idx := viper.GetString("index")
	is.BaseDir = filepath.Dir(idx)
	is.IndexFile = filepath.Base(idx)
	if verbose {
		fmt.Printf("[info] Getting index from %s\n", idx)
	}
	fp, err := openuri.Open(idx)
	if err != nil {
		return
	}
	data, err = ioutil.ReadAll(fp)
	if err != nil {
		return
	}
	var states []StateEntry
	err = json.Unmarshal(data, &states)
	if err != nil {
		if verbose {
			fmt.Printf("[error] Cannot unmarshall data, error: %s for data:\n%s", err.Error(), data)
		}
		return
	}
	is.Entries = states
	return
}

func OpenContent(url string) (io.ReadCloser, error) {
	fp, err := openuri.Open(url)
	return fp, err
}

func GetContent(url string) ([]byte, error) {
	fp, err := openuri.Open(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	key := viper.GetString("key")
	if key != "" {
		hash := md5.Sum([]byte(key))
		bdata, err := encrypt(string(data), []byte(hash[:]))
		if err != nil {
			return nil, err
		}
		data = []byte(bdata)
	}
	return data, nil
}

func GetMD5(url string) (sum string, err error) {
	var data []byte
	data, err = GetContent(url)
	if err != nil {
		return
	}
	sum = fmt.Sprintf("%x", md5.Sum(data))
	return
}

func VerifyIndexState(is *IndexState) (err error) {
	for k, v := range is.Entries {

		if v.Filename != "" {
			data, err := GetContent(is.BaseDir + "/" + v.Filename)
			if err != nil {
				return err
			}
			sum, err := GetMD5(is.BaseDir + "/" + v.Filename)
			if v.MD5SUM != "" && v.MD5SUM != sum {
				return fmt.Errorf("Bad MD5 sum for file %s", v.Filename)
			}
			is.Entries[k].MD5SUM = fmt.Sprintf("%x", md5.Sum(data))
		}
	}
	return nil
}
