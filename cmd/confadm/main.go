package main

import (
	"bufio"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/moensch/confclient"
	"io/ioutil"
	"os"
	"strings"
)

var (
	operation     string
	keyName       string
	configMgrUrl  string
	configMgrUser string
	configMgrPass string
	logLevel      string
)

func init() {
	flag.StringVar(&configMgrUser, "u", os.Getenv("CONFIGMGR_USER"), "Username")
	flag.StringVar(&configMgrPass, "p", os.Getenv("CONFIGMGR_PASS"), "Password")
	flag.StringVar(&configMgrUrl, "s", os.Getenv("CONFIGMGR_URL"), "Config manager URL")
	flag.StringVar(&logLevel, "l", "error", "Log level (debug|info|warn|error)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage for %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n\nOperation:\n")
		fmt.Fprintf(os.Stderr, "  get <key>                   : Get a key as JSON to stdout\n")
		fmt.Fprintf(os.Stderr, "  gett <key>                  : Get a key as TEXT to stdout\n")
		fmt.Fprintf(os.Stderr, "  geta <key-partial>          : Get all keys matching this partial key\n")
		fmt.Fprintf(os.Stderr, "  set <key>                   : Set a key from JSON read from stdin\n")
		fmt.Fprintf(os.Stderr, "  set <key> <value>           : Set a STRING key from parameter\n")
		fmt.Fprintf(os.Stderr, "  del <key>                   : Delete a key\n")
		fmt.Fprintf(os.Stderr, "  list <filter>               : List matching keys\n")
		fmt.Fprintf(os.Stderr, "  type <key>                  : Get Key Type\n")
		fmt.Fprintf(os.Stderr, "  hget <key> <field>          : Get just one field from a hash\n")
		fmt.Fprintf(os.Stderr, "  hset <key> <field> <value>  : Set just one field in a hash\n")
		fmt.Fprintf(os.Stderr, "  hgeta <key-partial> <field> : Get every place this key field is defined\n")
		fmt.Fprintf(os.Stderr, "  lget <key> <index>          : Get list item at position\n")
		fmt.Fprintf(os.Stderr, "  lpush <key> <value>         : Add entry to list (or create new list if it does not exist)\n")
		fmt.Fprintf(os.Stderr, "\n\nConfig:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	lvl, _ := log.ParseLevel(logLevel)
	log.SetLevel(lvl)
	if flag.NArg() < 2 {
		log.Warnf("Not enough args. Need 2, have %s", flag.NArg())
		flag.Usage()
		os.Exit(1)
	}
	operation = flag.Arg(0)
	keyName = flag.Arg(1)

	if operation == "" || configMgrUrl == "" || keyName == "" {
		log.Infof("Operation: %s", operation)
		log.Infof("configMgrUrl: %s", configMgrUrl)
		log.Infof("keyName: %s", keyName)
		flag.Usage()
		os.Exit(1)
	}
	var c = confclient.InitiateClient(configMgrUrl)

	switch operation {
	case "geta":
		keys, err := c.AdminListKeys("*:" + keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		for _, k := range keys {
			fmt.Printf("Key: %s\n", k)
			resp, err := c.AdminGetKeyAsTEXT(k)
			if err != nil {
				log.Fatalf("ERROR: %s", err)
			}
			fmt.Printf("%s\n", resp)
		}
	case "related":
		sl := strings.Split(keyName, ":")
		last_part := sl[len(sl)-1]
		keys, err := c.AdminListKeys("*:" + last_part)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		for _, k := range keys {
			fmt.Printf("%s\n", k)
		}
		log.Debug("LIST OK")
	case "list":
		keys, err := c.AdminListKeys(keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		for _, k := range keys {
			fmt.Printf("%s\n", k)
		}
		log.Debug("LIST OK")
	case "hlist":
		keys, err := c.AdminListHashFields(keyName)
		if err != nil {
			log.Debug("ERROR: %s", err)
		}
		for _, k := range keys {
			fmt.Printf("%s\n", k)
		}
		log.Debug("HLIST OK")
	case "del":
		err := c.AdminDeleteKey(keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		log.Debug("DEL OK")
	case "gett":
		stringresp, err := c.AdminGetKeyAsTEXT(keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		fmt.Printf("%s\n", stringresp)
		log.Debug("GETT OK")
	case "get":
		jsonblob, err := c.AdminGetKeyAsJSON(keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		fmt.Printf("%s\n", string(jsonblob))
		log.Print("GET OK")
	case "type":
		ktype, err := c.AdminGetKeyType(keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		fmt.Printf("%s\n", ktype)
		log.Debug("TYPE OK")
	case "set":
		if flag.NArg() >= 3 {
			// Someone also passed in a value
			value := flag.Arg(2)
			err := c.AdminSetStringKey(keyName, value)
			if err != nil {
				log.Fatalf("ERROR: %s", err)
			}
		} else {
			reader := bufio.NewReader(os.Stdin)
			b, err := ioutil.ReadAll(reader)
			if err != nil {
				log.Fatalf("ERROR: %s", err)
			}
			if len(b) == 0 {
				log.Fatalf("ERROR: Read zero bytes")
			}
			err = c.AdminSetKeyFromJSON(keyName, b)
			if err != nil {
				log.Fatalf("ERROR: %s", err)
			}
		}
		log.Debug("SET OK")
	case "hgeta":
		if flag.NArg() < 3 {
			flag.Usage()
			os.Exit(1)
		}
		fieldName := flag.Arg(2)
		keys, err := c.AdminListKeys("*:" + keyName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}

		found := make(map[string]string)
		maxlen := 0
		for _, k := range keys {
			val, err := c.AdminGetHashField(k, fieldName)
			if err != nil {
				// Ignore 404
				if !strings.Contains(err.Error(), "404") {
					// This is bad, string matching an error, but it gets shit done for now
					log.Fatalf("ERROR: Cannot get key %s field %s: %s", k, fieldName, err)
				} else {
					continue
				}
			}
			found[k] = val
			if len(k) > maxlen {
				maxlen = len(k)
			}
		}
		for k, v := range found {
			fmt.Printf("%s%s%s\n", k, strings.Repeat(" ", maxlen-len(k)+5), v)
		}
	case "hget":
		if flag.NArg() < 3 {
			flag.Usage()
			os.Exit(1)
		}
		fieldName := flag.Arg(2)

		val, err := c.AdminGetHashField(keyName, fieldName)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		fmt.Printf("%s\n", val)
		log.Debug("HGET OK")
	case "lpush":
		if flag.NArg() < 3 {
			flag.Usage()
			os.Exit(1)
		}
		stringval := flag.Arg(2)
		err := c.AdminListAppend(keyName, stringval)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
	case "hset":
		if flag.NArg() < 4 {
			flag.Usage()
			os.Exit(1)
		}
		fieldName := flag.Arg(2)
		stringval := flag.Arg(3)

		err := c.AdminSetHashField(keyName, fieldName, stringval)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		log.Debug("HSET OK")
	case "lget":
		if flag.NArg() < 3 {
			flag.Usage()
			os.Exit(1)
		}
		listIndex := flag.Arg(2)

		val, err := c.AdminGetListIndex(keyName, listIndex)
		if err != nil {
			log.Fatalf("ERROR: %s", err)
		}
		fmt.Printf("%s\n", val)
		log.Debug("LGET OK")
	}
}
