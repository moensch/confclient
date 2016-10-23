package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/moensch/confclient"
	"os"
	"strings"
	"text/template"
)

var (
	requestKey   string
	templateFile string
	logLevel     string
	configMgrUrl string
	verifyOutput bool
	configDir    string
	templateDir  string
)

func init() {
	flag.StringVar(&requestKey, "s", "", "Single variable mode")
	flag.StringVar(&templateFile, "t", "", "Template file (single template mode)")
	flag.StringVar(&configMgrUrl, "u", os.Getenv("CONFIGMGR_URL"), "Config manager URL")
	flag.BoolVar(&verifyOutput, "v", false, "Verify output")
	flag.StringVar(&logLevel, "l", "info", "Log level (debug|info|warn|error)")
	flag.StringVar(&configDir, "c", "/etc/conftpl/conf.d", "Configuration directory")
	flag.StringVar(&templateDir, "td", "/etc/conftpl/templates", "Template directory")
}

func main() {
	flag.Parse()
	lvl, _ := log.ParseLevel(logLevel)
	log.SetLevel(lvl)
	if configMgrUrl == "" {
		log.Fatal("configMgrUrl not set. Either set -u parameter or CONFIGMGR_URL environment variable")
	}
	var c = confclient.InitiateClient(configMgrUrl)
	c.TemplateDir = templateDir
	c.ConfigDir = configDir

	if requestKey != "" {
		// Want a single key
		val, err := c.GetString(requestKey)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
		}
		fmt.Printf("%s", val.Data.Value)
		os.Exit(0)
	}

	myFuncMap := template.FuncMap{
		"key":     c.GetStringValue,
		"keyd":    c.GetStringValueDebug,
		"list":    c.GetListValue,
		"listj":   c.GetListValueJoined,
		"listd":   c.GetListValueDebug,
		"hash":    c.GetHashValue,
		"hexists": c.HashExists,
		"sexists": c.StringExists,
		"lexists": c.ListExists,
	}
	if templateFile == "" {
		// No template provided on command line - read config
		templates, err := c.LoadConfigFiles()
		if err != nil {
			log.Fatalf("%s", err)
		}

		for _, t := range templates {
			log.Infof("Processing %s", t.Name)
			err := t.Process(myFuncMap)
			if err != nil {
				log.Fatalf("Error parsing template %s: %s", t.Src, err)
			}
			log.Infof("Successfully wrote: %s", t.Dest)
		}
		os.Exit(0)
	} else {
		// Processing single template from command line - printing to STDOUT
		tc := confclient.TemplateConfig{}
		tc.Src = templateFile
		if verifyOutput {
			if flag.NArg() < 1 {
				log.Fatal("ERROR: Must provide extra params for verify command")
			}

			tc.CheckCmd = strings.Join(flag.Args(), " ")
		}
		err := tc.Process(myFuncMap)
		if err != nil {
			log.Fatalf("Error parsing template %s: %s", templateFile, err)
		}
		os.Exit(0)
	}
}
