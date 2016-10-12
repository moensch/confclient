package main

import (
	"bytes"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/moensch/confclient"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
	"text/template"
)

var (
	requestKey   string
	templateFile string
	logLevel     string
	configMgrUrl string
	verifyOutput bool
)

func init() {
	flag.StringVar(&requestKey, "s", "", "Single variable mode")
	flag.StringVar(&templateFile, "t", "", "Template file")
	flag.StringVar(&configMgrUrl, "u", os.Getenv("CONFIGMGR_URL"), "Config manager URL")
	flag.BoolVar(&verifyOutput, "v", false, "Verify output")
	flag.StringVar(&logLevel, "l", "info", "Log level (debug|info|warn|error")
}

func main() {
	flag.Parse()
	lvl, _ := log.ParseLevel(logLevel)
	log.SetLevel(lvl)
	if configMgrUrl == "" {
		log.Fatal("configMgrUrl not set. Either set -u parameter or CONFIGMGR_URL environment variable")
	}
	var c = confclient.InitiateClient(configMgrUrl)

	if requestKey != "" {
		// Want a single key
		val, err := c.GetString(requestKey)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
		}
		fmt.Printf("%s", val.Data.Value)
		os.Exit(0)
	}
	if templateFile != "" {
		log.WithFields(log.Fields{"file": templateFile}).Info("Processing template")
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
		tmpl, err := template.New(path.Base(templateFile)).Funcs(myFuncMap).ParseFiles(templateFile)
		if err != nil {
			log.Fatalf("Error parsing template %s: %s", templateFile, err)
		}

		var buffer bytes.Buffer
		if err = tmpl.Execute(&buffer, nil); err != nil {
			log.Fatalf("Cannot execute template: %s", err)
		}

		if verifyOutput {
			// Write template to a temp file
			tmpfile, err := ioutil.TempFile("/tmp", "conftpl")
			if err != nil {
				log.Fatalf("Cannot create temp file: %s", err)
			}
			log.WithFields(log.Fields{
				"tempfile": tmpfile.Name(),
			}).Info("Writing to temp file")
			if _, err := tmpfile.Write(buffer.Bytes()); err != nil {
				log.Fatalf("Cannot write to temp file %s: %s", tmpfile.Name(), err)
			}
			if err := tmpfile.Close(); err != nil {
				log.Fatalf("Error closing temp file: %s", err)
			}

			// Making sure we have at least one extra argument for the verify command
			if flag.NArg() < 1 {
				log.Fatal("ERROR: Must provide extra params for verify command")
			}

			verify_cmd := flag.Arg(0)
			verify_args := make([]string, flag.NArg()-1)
			for i := 1; i < flag.NArg(); i++ {
				a := flag.Arg(i)
				if a == "FILE" {
					a = tmpfile.Name()
				}
				verify_args[i-1] = a
			}
			log.Printf("Verify Command: %s", verify_cmd)
			for _, a := range verify_args {
				log.Infof("  %s", a)
			}
			cmd := exec.Command(verify_cmd, verify_args...)
			if err := cmd.Start(); err != nil {
				log.Fatalf("cmd.Start: %v")
			}
			if err := cmd.Wait(); err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						log.Fatalf("Verify Exit Status: %d", status.ExitStatus())
					}
				} else {
					log.Fatalf("Verify cmd.Wait: %v", err)
				}
			}
			os.Remove(tmpfile.Name())
		}

		// Print template to stdout
		fmt.Printf("%s", buffer.String())

		os.Exit(0)
	}

	flag.Usage()
	os.Exit(1)
}
