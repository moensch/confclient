package main

import (
	"flag"
	"fmt"
	"github.com/moensch/confclient"
	"log"
	"os"
	"path"
	"text/template"
)

var (
	requestKey   string
	templateFile string
	showDebug    bool
	configMgrUrl string
)

func init() {
	flag.StringVar(&requestKey, "v", "", "Variable to request")
	flag.StringVar(&templateFile, "t", "", "Template file")
	flag.StringVar(&configMgrUrl, "u", os.Getenv("CONFIGMGR"), "Config manager URL")
	flag.BoolVar(&showDebug, "d", false, "Enable Debug output")
}

func main() {
	flag.Parse()
	var c = confclient.InitiateClient(configMgrUrl)

	if requestKey != "" {
		// Want a single key
		val, err := c.GetStringValue(requestKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf(val)
		os.Exit(0)
	}
	if templateFile != "" {
		log.Printf("Processing template: %s", templateFile)
		myFuncMap := template.FuncMap{
			"key":   c.GetStringValue,
			"keyd":  c.GetStringValueDebug,
			"list":  c.GetListValue,
			"listd": c.GetListValueDebug,
			"hash":  c.GetHashValue,
		}
		tmpl, err := template.New(path.Base(templateFile)).Funcs(myFuncMap).ParseFiles(templateFile)
		if err != nil {
			log.Printf("Error parsing template %s: %s", templateFile, err)
		}

		if err = tmpl.Execute(os.Stdout, nil); err != nil {
			log.Printf("Cannot execute template: %s", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	flag.Usage()
	os.Exit(1)
}
