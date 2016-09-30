# confclient

[![Build Status](https://travis-ci.org/moensch/confclient.svg?branch=master)](https://travis-ci.org/moensch/confclient)

Example client implementation and template parser for confmgr

## confadm Usage

## conftpl Usage

### Get string variable
```
conftpl -u http://confmgr:8080 -s somestringkey
conftpl -u http://confmgr:8080 -s somehash/field
conftpl -u http://confmgr:8080 -s somelist/index/0
```

### Template processing
```
conftpl -u http://confmgr:8080 -t my_template.tmpl > my_template.output
```

### With verification

Note: The "FILE" parameter will be replaced with the parsed template output file

```
conftpl -u http://confmgr:8080 -t some_xml.tmpl -v -- xmllint --format FILE > some_xml.xml
```

## Building

```
go get github.com/moensch/confmgr/cmd/conftpl
go get github.com/moensch/confmgr/cmd/confadm
```

# Template functions

* key "keyName" "defaultValue"
  * Gets a given key, optionally fall back to default value
  * Use "keyName/field" to retrieve hash fields
  * Use "keyName/index/n" to retrieve a list item
* keyd "keyName" "defaultValue"
  * Same as the above, but it will return a hash with "Value" and "Source" to show value origin
* list "keyName"
  * Retrieve a list
* listd "keyName"
  * Same as the above, but each list item will be a hash with "Value" and "Source"
* listj "keyName" "joinChar"
  * Retrieve a list, but return as a string, joined by "joinChar"
* hash "keyName"
  * Retrieve a hash
  * Always sets "Source" hash key


## Examples

### key

```
Hello: {{key "name" "Stranger"}}
```

### keyd

```
{{with keyd "name" "Stranger"}}
Hello {{.Value}} (key stored at {{.Source}}
{{end}}
```

### list

```
{{range list "people"}}
  Hello {{.}}
{{end}}
```

### listd

```
{{range listd "people"}}
  Hello {{.Value}} (From: {{.Source}})
{{end}}
```

### listj

```
I welcome the following people: {{listj "people" ", "}}
```

### hash

```
{{range hash "db_settings"}}
  <!-- From: {{.Source}} -->
  <attr key='{{.Key}}' val='{{.Value}}'/>{{end}}
```
