package ping_simple

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
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
	bytes, _ = ioutil.ReadFile("setting.json")
	var waitTime int
	_ = json.Unmarshal(bytes, &waitTime)
	hosts_data := ini(hosts)

	for {
		c := exec.Command("cmd", "/c", "cls")
		c.Stdout = os.Stdout
		c.Run()
		for _, h := range hosts_data {
			go Ping_inside_simple(h.host, h.c)
		}
		for _, h := range hosts_data {
			re := <-h.c
			fmt.Printf("%-20s %-6.2fms %-4.2f%%\n", h.host, re.Average, re.LostRate*100)
		}
		temp_symbol := [4]string{"↑", "→", "↓", "←"}
		print("wait1 ")
		for i := 0; i < waitTime*2; i++ {
			fmt.Printf("\b\b\b\b\b\bwait%s", temp_symbol[i%4])
			time.Sleep(time.Duration(1) * time.Millisecond * 500)
		}

	}
}
