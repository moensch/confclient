package confclient

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"
)

type TemplateConfigFile struct {
	Tpl TemplateConfig `toml:"template"`
}

type TemplateConfig struct {
	Src      string `toml:"src"`
	Name     string
	Dest     string `toml:"dest"`
	TempDest string
	Uid      int `toml:"uid"`
	Gid      int `toml:"gid"`
	FileMode os.FileMode
	Mode     string `toml:"mode"`
	CheckCmd string `toml:"check_cmd"`
}

func (t *TemplateConfig) Process(funcMap map[string]interface{}) error {
	log.WithFields(log.Fields{"file": t.Src}).Info("Processing template")
	tmpl, err := template.New(filepath.Base(t.Src)).Funcs(funcMap).ParseFiles(t.Src)
	if err != nil {
		return fmt.Errorf("Error parsing template %s: %s", t.Src, err)
	}
	var buffer bytes.Buffer
	if err = tmpl.Execute(&buffer, nil); err != nil {
		return fmt.Errorf("Cannot execute template %s: %s", t.Src, err)
	}

	var tmpfile *os.File
	if t.TempDest != "" {
		tmpfile, err = os.OpenFile(t.TempDest, os.O_WRONLY|os.O_CREATE, t.FileMode)
		if err != nil {
			return fmt.Errorf("Cannot open temporary file %s: %s", t.TempDest, err)
		}
	} else {
		tmpfile, err = ioutil.TempFile("/tmp", filepath.Base(t.Src))
		if err != nil {
			return fmt.Errorf("Cannot create temporary file: %s", err)
		}
		t.TempDest = tmpfile.Name()
	}
	if _, err := tmpfile.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("Cannot write to temp file %s: %s", tmpfile.Name(), err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("Error closing temp file: %s", err)
	}

	if t.CheckCmd != "" {
		err := t.Validate()
		if err != nil {
			os.Remove(t.TempDest)
			return fmt.Errorf("Validation failed for %s: %s", t.Src, err)
		}
	}

	if t.Dest == "" {
		// No destination - spew to stdout
		os.Remove(tmpfile.Name())
		fmt.Printf("%s", buffer.String())
		return nil
	} else {
		// Rename to destination file
		os.Chmod(t.TempDest, t.FileMode)
		os.Chown(t.TempDest, t.Uid, t.Gid)
		err := os.Rename(t.TempDest, t.Dest)
		if err != nil {
			return fmt.Errorf("Cannot rename %s to %s: %s", t.TempDest, t.Dest, err)
		}
	}
	return nil
}

func (t *TemplateConfig) Validate() error {
	if t.CheckCmd == "" {
		return nil
	}
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = t.TempDest
	tmpl, err := template.New("checkcmd").Parse(t.CheckCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	log.Debug("Running " + cmdBuffer.String())
	c := exec.Command("/bin/sh", "-c", cmdBuffer.String())
	/*
		mycmd := strings.Replace(t.CheckCmd, "FILE", t.TempDest, -1)
		log.Infof("Running verify: %s", mycmd)
		c := exec.Command("/bin/sh", "-c", mycmd)
	*/
	output, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\nOutput: %s", err, string(output))
	}
	log.Debugf("%q", string(output))
	return nil
}

type TemplateConfigs []*TemplateConfig

func (c *Client) LoadConfigFiles() (TemplateConfigs, error) {
	log.Infof("Reading config from: '%s'", c.ConfigDir)

	var tcs TemplateConfigs
	dh, err := os.Open(c.ConfigDir)
	if err != nil {
		return tcs, fmt.Errorf("Cannot open dir %s: %s", c.ConfigDir, err)
	}

	files, err := dh.Readdir(-1)
	if err != nil {
		return tcs, fmt.Errorf("Cannot read dir %s: %s", c.ConfigDir, err)
	}
	for _, fi := range files {
		tc, err := c.LoadConfigFile(filepath.Join(c.ConfigDir, fi.Name()))
		if err != nil {
			return nil, fmt.Errorf("Cannot load config file %s: %s", filepath.Join(c.ConfigDir, fi.Name()), err)
		}
		tcs = append(tcs, tc)
	}

	return tcs, nil
}

func (c *Client) LoadConfigFile(path string) (*TemplateConfig, error) {
	tc := &TemplateConfigFile{TemplateConfig{Uid: -1, Gid: -1}}
	log.Infof("Loading config %s", path)
	_, err := toml.DecodeFile(path, &tc)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse config %s: %s", path, err)
	}

	tr := tc.Tpl
	tr.Name = tr.Src
	if tr.Uid == -1 {
		tr.Uid = os.Geteuid()
	}
	if tr.Gid == -1 {
		tr.Gid = os.Getegid()
	}
	if tr.Mode != "" {
		mode, err := strconv.ParseUint(tr.Mode, 0, 32)
		if err != nil {
			return nil, err
		}
		tr.FileMode = os.FileMode(mode)
	}
	tr.Src = filepath.Join(c.TemplateDir, tr.Src)
	tr.TempDest = fmt.Sprintf("%s.tmp", tr.Dest)

	return &tr, nil
}
