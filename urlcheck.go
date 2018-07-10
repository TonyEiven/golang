package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	//"sync"
	"flag"
)

const conf = "eureka.json"

var (
	File_error    = errors.New("Could't find config file")
	Perm_error    = errors.New("Need permission to access the file")
	Request_error = errors.New("Request failed")
	help          bool
	version       bool
	webhook       string
)

func init() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&version, "v", false, "show version and exit")
	flag.StringVar(&webhook, "w", "", "`robot hook`")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `url_check version: url_check/1.0 Usage: url_check [-?hvw] [-w webhook]
				Options:`)
	flag.PrintDefaults()
}

func FileExist() bool {
	currentPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	_, err := os.Stat(currentPath + "/" + conf)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		fmt.Println(File_error)
		return false
	}
	return false
}

type HealthRes struct {
	Hystrix struct {
		Status string `json:"status"`
	} `json:"hystrix"`
	ConfigServer struct {
		PropertySources []string `json:"propertySources"`
		Status          string   `json:"status"`
	} `json:"configServer"`
	Status             string `json:"status"`
	DiscoveryComposite struct {
		Eureka struct {
			Applications struct {
				MESSAGEPROCESSORWECHAT      int `json:"MESSAGE-PROCESSOR-WECHAT"`
				MESSAGEPROCESSORSMS         int `json:"MESSAGE-PROCESSOR-SMS"`
				FOUNDATIONSYNC              int `json:"FOUNDATION-SYNC"`
				FOUNDATIONINVITE            int `json:"FOUNDATION-INVITE"`
				FOUNDATIONCAPTCHA           int `json:"FOUNDATION-CAPTCHA"`
				FOUNDATIONTOKEN             int `json:"FOUNDATION-TOKEN"`
				MONITOR                     int `json:"MONITOR"`
				FOUNDATIONTENANT            int `json:"FOUNDATION-TENANT"`
				MESSAGEPROCESSORPUSH        int `json:"MESSAGE-PROCESSOR-PUSH"`
				MESSAGETIMEDTASK            int `json:"MESSAGE-TIMED-TASK"`
				FOUNDATIONSYSTEM            int `json:"FOUNDATION-SYSTEM"`
				MESSAGEMANAGER              int `json:"MESSAGE-MANAGER"`
				FOUNDATIONAUTH              int `json:"FOUNDATION-AUTH"`
				MESSAGEGATEWAY              int `json:"MESSAGE-GATEWAY"`
				FOUNDATIONGATEWAY           int `json:"FOUNDATION-GATEWAY"`
				MESSAGEPROCESSORINTERNALMSG int `json:"MESSAGE-PROCESSOR-INTERNAL-MSG"`
				FOUNDATIONUSER              int `json:"FOUNDATION-USER"`
				CONFIG                      int `json:"CONFIG"`
				FOUNDATIONENCRYPT           int `json:"FOUNDATION-ENCRYPT"`
				MESSAGEADMIN                int `json:"MESSAGE-ADMIN"`
				FOUNDATIONNOTIFY            int `json:"FOUNDATION-NOTIFY"`
			} `json:"applications"`
			Status      string `json:"status"`
			Description string `json:"description"`
		} `json:"eureka"`
		DiscoveryClient struct {
			Services    []string `json:"services"`
			Status      string   `json:"status"`
			Description string   `json:"description"`
		} `json:"discoveryClient"`
		Status      string `json:"status"`
		Description string `json:"description"`
	} `json:"discoveryComposite"`
	Mail struct {
		Error    string `json:"error"`
		Location string `json:"location"`
		Status   string `json:"status"`
	} `json:"mail"`
	DiskSpace struct {
		Threshold int    `json:"threshold"`
		Free      int64  `json:"free"`
		Total     int64  `json:"total"`
		Status    string `json:"status"`
	} `json:"diskSpace"`
	Rabbit struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	} `json:"rabbit"`
	Redis struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	} `json:"redis"`
	Db struct {
		Hello    int    `json:"hello"`
		Database string `json:"database"`
		Status   string `json:"status"`
	} `json:"db"`
	RefreshScope struct {
		Status string `json:"status"`
	} `json:"refreshScope"`
}

type ConfParse struct {
	Apps []struct {
		Name      string   `json:"Name"`
		InstaceID []string `json:"InstaceID"`
		HealthURL []string `json:"HealthURL"`
	} `json:"apps"`
}

type Body struct {
	Msgtype  string   `json:"msgtype"`
	Markdown Markdown `json:"markdown"`
}

type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func UrlAvail(uri string) bool {
	_, err := http.Get(uri)
	if err != nil {
		fmt.Println(Request_error, err)
		return false
	}
	return true
}

func Request(job chan string, result chan []byte) {
	for j := range job {
		res, err := http.Get(j)
		if err != nil {
			fmt.Println(Request_error, err)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Read response body failed!~ %s", err)
		}
		result <- body
	}
}

func (h *HealthRes) GetStatus() string {
	return h.Status
}
func (h *HealthRes) GetConfServerStatus() string {
	return h.ConfigServer.Status
}
func (h *HealthRes) GetHystrixStatus() string {
	return h.Hystrix.Status
}

func (h *HealthRes) GetDBStatus() string {
	return h.Db.Status
}

func (h *HealthRes) GetRedisStatus() string {
	return h.Redis.Status
}

func (h *HealthRes) GetRabbitStatus() string {
	return h.Rabbit.Status
}

func DingNotify(webhook string, appname string, instanceID string) (string, bool) {
	var b bool
	var op string
	var data string = fmt.Sprintf("Severity: Severe!!! MsgObject: One of Eureka application %s instance %s is DOWN", appname, instanceID)
	v := Body{Msgtype: "markdown", Markdown: Markdown{Title: "Warning", Text: data}}
	res, _ := json.Marshal(v)
	client := http.Client{}
	req, _ := http.NewRequest("POST", webhook, bytes.NewBuffer(res))
	req.Header.Set("Content-Type", "application/json")

	rep, _ := client.Do(req)
	defer rep.Body.Close()
	if rep.StatusCode == 200 {
		body, _ := ioutil.ReadAll(rep.Body)
		op = string(body)
		b = true
	} else {
		body, _ := ioutil.ReadAll(rep.Body)
		op = string(body)
		b = false
	}
	return op, b
}
func main() {
	flag.Parse()
	if os.Args[1] == "-h" || len(os.Args) == 1 {
		usage()
	}
	args := flag.Args()
	if len(args) <= 1 {
		usage()
	}
	//var wg sync.WaitGroup
	var job = make(chan string, 1)
	var result = make(chan []byte, 1)
	if FileExist() == true {
		rf, err := os.Open(conf)
		if err != nil {
			fmt.Println(Perm_error, err)
		}
		ct, _ := ioutil.ReadAll(rf)
		pc := ConfParse{}
		hrr := HealthRes{}
		err = json.Unmarshal(ct, &pc)
		if err != nil {
			fmt.Println(err)
		}
		for _, apps := range pc.Apps {
			for c, ins := range apps.HealthURL {
				Res := UrlAvail(ins)
				if Res != true {
					DingNotify(webhook, apps.Name, apps.InstaceID[c])
				}
				job <- ins
				go Request(job, result)
				r := <-result
				err = json.Unmarshal(r, &hrr)
				if err != nil {
					fmt.Println(err)
				}
				if hrr.GetDBStatus() == "DOWN" || hrr.GetRabbitStatus() == "DOWN" || hrr.GetRedisStatus() == "DOWN" {
					DingNotify(webhook, apps.Name, apps.InstaceID[c])
				}
				fmt.Println(apps.Name, apps.InstaceID, hrr.GetStatus(), hrr.GetRedisStatus(), hrr.GetConfServerStatus(), hrr.GetRabbitStatus())
			}
		}
	}
}
