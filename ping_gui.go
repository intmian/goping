package ping_simple

import (
	"encoding/json"
	"io/ioutil"
)

type host_info struct {
	host string
	c    chan Ping_info
}

func ini(hosts []string) []host_info {
	var result []host_info

	for _, host := range hosts {
		result = append(result, host_info{host, make(chan Ping_info)})
	}
	return result
}

func Gui() {
	bytes, _ := ioutil.ReadFile("hosts.json")
	var hosts []string
	_ = json.Unmarshal(bytes, &hosts)
	bytes, _ = ioutil.ReadFile("hosts.json")
	var waitTime int
	_ = json.Unmarshal(bytes, &waitTime)

}
