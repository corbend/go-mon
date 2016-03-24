package daemon

import (
	"fmt"
	"log"
	"bufio"
	"os/exec"
	"flag"
	"strconv"
	"net/http"
	"strings"
	"regexp"
	"html/template"	
)

type DaemonConfig struct {
	port int
}

type DiskStat struct {
	Filesystem string
	Size string
	Used string
	Avail string
	Use string
	Mounted string
}

type CpuStat struct {
	Uptime string
	LoadAvg1 string
	LoadAvg2 string
	LoadAvg3 string	
}

type MemStat struct {
	Swpd string
	Free string
	Buff string
	Cache string
	SwapSi string
	SwapSo string
	IoBi string
	IoBo string
}

type TrafStat struct {

}

var monChanDisk = make(chan []DiskStat)
var monChanCpu = make(chan []CpuStat)
var monChanMem = make(chan []MemStat)
var monChanTraf = make(chan []TrafStat)

func ParseDisk() {

	_, err := exec.LookPath("df")
	if err != nil {
		log.Fatalln("error on looking command")
	}

	resp, err := exec.Command("df", "-Ph").Output()

	fmt.Printf("resp - \r\n%s\r\n", resp)

	s := bufio.NewScanner(strings.NewReader(string(resp)))

	cnt := 0
	re := regexp.MustCompile(`\s+`)

	rows := []DiskStat{}

	for s.Scan() {
		line := s.Text()
		cnt += 1

		if cnt > 1 {
			df := DiskStat{}
			columns := re.Split(line, -1)

			for c, data := range columns {
				if c == 0 {
					df.Filesystem = data
				} else if c == 1 {
					df.Size = data
				} else if c == 2 {
					df.Used = data
				} else if c == 3 {
					df.Avail = data
				} else if c == 4 {
					df.Use = data
				} else if c == 5 {
					df.Mounted = data
				}
			}
			rows = append(rows, df)
			fmt.Printf("line = %s\r\n", df)
		}
	}

	monChanDisk <- rows
}

func ParseCpu() {

	_, err := exec.LookPath("cat")
	if err != nil {
		log.Fatalln("error on looking command")
	}

	_, err = exec.Command("cat", "/proc/loadavg").Output()

	rows := make([]CpuStat, 1)

	monChanCpu <- rows
}

func ParseMem() {

	_, err := exec.LookPath("vmstat")
	if err != nil {
		log.Fatalln("error on looking command")
	}

	re := regexp.MustCompile(`\s+`)

	resp, err := exec.Command("vmstat").Output()

	s := bufio.NewScanner(strings.NewReader(string(resp)))

	rows := []MemStat{}

	cnt := 0

	for s.Scan() {
		line := s.Text()
		cnt += 1

		if cnt > 2 {
			mf := MemStat{}
			columns := re.Split(line, -1)

			for c, data := range columns {
				if c == 3 {
					mf.Swpd = data
				} else if c == 4 {
					mf.Free = data
				} else if c == 5 {
					mf.Buff = data
				} else if c == 6 {
					mf.Cache = data
				} else if c == 7 {
					mf.SwapSi = data
				} else if c == 8 {
					mf.SwapSo = data
				} else if c == 9 {
					mf.IoBi = data
				} else if c == 10 {
					mf.IoBo = data
				}
			}
			rows = append(rows, mf)
			fmt.Printf("line = %s\r\n", mf)
		}
	}

	monChanMem <- rows
}

func ParseTraf() {

	_, err := exec.LookPath("df")
	if err != nil {
		log.Fatalln("error on looking command")
	}

	_, err = exec.Command("vmstat").Output()

	rows := make([]TrafStat, 1)

	monChanTraf <- rows
}

type Output struct {
	Name string
	Disk []DiskStat
	Mem []MemStat
	Cpu []CpuStat
	Traf []TrafStat
}

func makeHandler() func(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("stat_view.html")

	if err != nil {
		log.Fatalf("template error")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		go ParseDisk()
		go ParseMem()
		go ParseCpu()
		go ParseTraf()

		diskInfo := <- monChanDisk
		memInfo := <- monChanMem
		cpuInfo := <- monChanCpu
		trafInfo := <- monChanTraf

		out := Output{}
		out.Name = "SERVER INFO"
		out.Disk = diskInfo
		out.Mem = memInfo
		out.Cpu = cpuInfo
		out.Traf = trafInfo

		t.Execute(w, &out)
	}
}

func Start() {

	var port int

	flag.IntVar(&port, "p", 8080, "port for daemon")
	flag.Parse()

	conf := DaemonConfig{}
	conf.port = port

	http.HandleFunc("/", makeHandler())
    http.ListenAndServe(":" + strconv.Itoa(conf.port), nil)
}