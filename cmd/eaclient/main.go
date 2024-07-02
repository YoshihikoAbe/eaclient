package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/YoshihikoAbe/avsproperty"
	"github.com/YoshihikoAbe/eaclient"
	"gopkg.in/yaml.v3"
)

var (
	userAgent, srcid, model string

	config struct {
		Client   eaclient.Client             `yaml:"client"`
		Services map[string]eaclient.Service `yaml:"services"`
	}
)

func main() {
	flag.StringVar(&userAgent, "u", "", "Override the value of the client's User-Agent header")
	flag.StringVar(&srcid, "p", "", "Override the client's srcid")
	flag.StringVar(&model, "m", "", "Override the client's model")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] CONFIG SERVICE REQUEST\nList of available options:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) < 3 {
		flag.Usage()
		os.Exit(1)
	}

	if err := loadConfig(); err != nil {
		fatal("failed to load config:", err)
	}
	transport := http.DefaultTransport.(*http.Transport)
	transport.DisableCompression = true
	transport.ForceAttemptHTTP2 = false

	svcName := flag.Arg(1)
	svc, ok := config.Services[flag.Arg(1)]
	if !ok {
		fatal("service not found:", svcName)
	}

	prop := &avsproperty.Property{}
	if err := prop.ReadFile(flag.Arg(2)); err != nil {
		fatal("failed to read property:", err)
	}
	resp, err := config.Client.Send(svc, prop.Root)
	if err != nil {
		fatal(err)
	}

	resp.Settings.Format = avsproperty.FormatPrettyXML
	if err := resp.Write(os.Stdout); err != nil {
		fatal(err)
	}
}

func loadConfig() error {
	b, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return err
	}

	if userAgent != "" {
		config.Client.UserAgent = userAgent
	}
	if srcid != "" {
		config.Client.Srcid = srcid
	}
	if model != "" {
		config.Client.Model = model
	}
	return nil
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
