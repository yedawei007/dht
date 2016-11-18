package main

import (
	"encoding/hex"
	"encoding/json"
	//"fmt"
	//"github.com/shiyanhui/dht"
	"dht"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"flag"
)

type file struct {
	Path   []interface{} `json:"path"`
	Length int           `json:"length"`
}

type bitTorrent struct {
	InfoHash string `json:"infohash"`
	Name     string `json:"name"`
	Files    []file `json:"files,omitempty"`
	Length   int    `json:"length,omitempty"`
}

func main() {
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	url := flag.String("url","", "sprider -url=http://localhost:8080/sendData");
	flag.Parse()
	ch := make(chan string, 100)
	go RunSave(ch, *url)
	w := dht.NewWire(65536, 1024, 256)
	go func() {
		for resp := range w.Response() {
			metadata, err := dht.Decode(resp.MetadataInfo)
			if err != nil {
				continue
			}
			info := metadata.(map[string]interface{})

			if _, ok := info["name"]; !ok {
				continue
			}

			bt := bitTorrent{
				InfoHash: hex.EncodeToString(resp.InfoHash),
				Name:     info["name"].(string),
			}

			if v, ok := info["files"]; ok {
				files := v.([]interface{})
				bt.Files = make([]file, len(files))

				for i, item := range files {
					f := item.(map[string]interface{})
					bt.Files[i] = file{
						Path:   f["path"].([]interface{}),
						Length: f["length"].(int),
					}
				}
			} else if _, ok := info["length"]; ok {
				bt.Length = info["length"].(int)
			}

			data, err := json.Marshal(bt)
			if err == nil {
				//fmt.Printf("%s\n\n", data)
				ch <- string(data)
			}
		}
	}()
	go w.Run()

	config := dht.NewCrawlConfig()
	config.OnAnnouncePeer = func(infoHash, ip string, port int) {
		w.Request([]byte(infoHash), ip, port)
	}
	d := dht.New(config)

	d.Run()
}


func RunSave(c chan string, url string) {
	client := http.Client{}
	for true {
		s := <-c
		req, _ := http.NewRequest("POST", url, strings.NewReader(s))
		_, _ = client.Do(req)
	}
}

