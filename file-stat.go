package main

import (
	log "github.com/sirupsen/logrus"
	"crypto/tls"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"strconv"
)

var (
	pushgateway  string
	pushinterval time.Duration
	start        time.Time
	InsecureSkipVerify bool
)

func init() {
	var err error
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05",
		ForceColors:            true,
		DisableLevelTruncation: true,
	})
	envs := []string{"PUSHGATEWAY_URL", "DIR1_PATH"}
	check_crtitical_envs(envs)

	pushgateway = os.Getenv("PUSHGATEWAY_URL")
	if os.Getenv("TLS_SKIP_VERIFY") == "" {
		InsecureSkipVerify = false
	}
	InsecureSkipVerify, _ = strconv.ParseBool(os.Getenv("TLS_SKIP_VERIFY"))
	pushintervalstring := os.Getenv("PUSH_INTERVAL")
	if pushintervalstring == "" {
		pushintervalstring = "60s"
	}
	pushinterval, err = time.ParseDuration(pushintervalstring)
	if err != nil {
		log.Fatal("Wrong value of " + pushintervalstring + ", error: " + err.Error())
	}
}

func check_crtitical_envs(envs []string) {
	for _, key := range envs {
		if _, ok := os.LookupEnv(key); !ok {
			log.Fatal("Environment " + key + " not set!")
		}

	}
}

type metric struct {
	label string
	size  int64
}

func read_config() {
	var (
		label string
		path  string
		ext   string
	)
	log.Info("Collecting metrics...")
	start = time.Now()
	for _, element := range os.Environ() {
		variable := strings.Split(element, "=")
		if strings.HasPrefix(variable[0], "DIR") && strings.HasSuffix(variable[0], "PATH") {
			dir := strings.Split(variable[0], "_")
			for _, element := range os.Environ() {

				variableInside := strings.Split(element, "=")
				if variableInside[0] == dir[0]+"_PATH" {
					path = variableInside[1]
				}
				if variableInside[0] == dir[0]+"_LABEL" {
					label = variableInside[1]
				}
				if variableInside[0] == dir[0]+"_EXT" {
					ext = variableInside[1]
				}
			}
			read_files(path, ext, label)
		}
	}
}
func visit(files *[]os.FileInfo, ext string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if filepath.Ext(path) == "."+ext {
			// log.Println(info)
			*files = append(*files, info)
		}

		return nil
	}
}

func read_files(root, ext, label string) {
	var files []os.FileInfo
	err := filepath.Walk(root, visit(&files, ext))
	if err != nil {
		log.Error(err)
	}
	push_metrics(&files, label)
}

func push_metrics(files *[]os.FileInfo, label string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: InsecureSkipVerify},
	}
	// client := &http.Client{Transport: tr}

	completionSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "file_stat_size_bytes",
		Help: "The size of file in a specific directory with additional information in labels",
	})
	for _, file := range *files {

		completionSize.Set(float64(file.Size()))
		if err := push.New(pushgateway, "file_stat").
			Collector(completionSize).
			Grouping("dir_label", label).
			Grouping("name", file.Name()).
			Grouping("modtime", file.ModTime().String()).
			Client(&http.Client{Transport: tr}).
			Push(); err != nil {
			log.Error("Could not push completion time to Pushgateway:", err)
		}

	}
	log.Info("Done, took: " + time.Since(start).String())
}
func main() {
	boolPtr := flag.Bool("run", false, "Run server")

	flag.Parse()
	if *boolPtr {
		for {

			read_config()
			time.Sleep(pushinterval)
		}
	} else {
		flag.PrintDefaults()
		log.Info(`
		Available environment variables:
			PUSHGATEWAY_URL - required
			PUSH_INTERVAL  - default 60s
			DIR{0-9}_PATH  - first required, e.g. DIR1_PATH
			DIR{0-9}_LABEL - first required, e.g. DIR1_LABEL
			DIR{0-9}_EXT   - first required, e.g. DIR1_EXT
		`)
	}

}
