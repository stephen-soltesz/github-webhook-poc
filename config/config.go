package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/m-lab/go/rtx"
	"github.com/stephen-soltesz/pretty"
)

// Repos encodes the list of allowed github repo HTML URLs. All others will be
// ignored.
type Repos map[string]string

// Load downloads a Repos config from the given URL.
func Load(configURL string) Repos {
	r := Repos{}
	log.Println("loading:", configURL)
	resp, err := http.Get(configURL)
	if err != nil {
		log.Println(err)
		return nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	rtx.Must(err, "Failed to read config body")
	json.Unmarshal(b, &r)
	pretty.Print(r)
	return r
}