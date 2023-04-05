package domains

import (
	"os"
	"path/filepath"
	"sqlow/database"
	"sqlow/helpers"
	"strings"
)

type Migration struct {
	Description string `yaml:"description"`

	Check       []string `yaml:"check"`
	OnFail      []string `yaml:"onFail"`
	OnSuccess   []string `yaml:"onSuccess"`
	OnNoResults []string `yaml:"onNoResults"`
	OnResults   []string `yaml:"onResults"`

	CheckFile       string `yaml:"checkFile"`
	OnFailFile      string `yaml:"onFailFile"`
	OnSuccessFile   string `yaml:"onSuccessFile"`
	OnNoResultsFile string `yaml:"onNoResultsFile"`
	OnResultsFile   string `yaml:"onResultsFile"`
}

func (c *Migration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Description = getValue("description", unmarshal)[0]

	c.Check = getValue("check", unmarshal)
	c.OnFail = getValue("onFail", unmarshal)
	c.OnSuccess = getValue("onSuccess", unmarshal)
	c.OnNoResults = getValue("onNoResults", unmarshal)
	c.OnResults = getValue("onResults", unmarshal)

	if checkFile := getValue("checkFile", unmarshal); len(checkFile) > 0 {
		c.CheckFile = checkFile[0]
	}
	if onFailFile := getValue("onFailFile", unmarshal); len(onFailFile) > 0 {
		c.OnFailFile = onFailFile[0]
	}
	if onSuccessFile := getValue("onSuccessFile", unmarshal); len(onSuccessFile) > 0 {
		c.OnSuccessFile = onSuccessFile[0]
	}
	if onNoResultsFile := getValue("onNoResultsFile", unmarshal); len(onNoResultsFile) > 0 {
		c.OnNoResultsFile = onNoResultsFile[0]
	}
	if onResultsFile := getValue("onResultsFile", unmarshal); len(onResultsFile) > 0 {
		c.OnResultsFile = onResultsFile[0]
	}
	return nil
}

func (c *Migration) RunCheck(driver database.Driver) (bool, bool) {
	for _, check := range c.Check {
		ok, hasRows := driver.QueryPasses(check)
		if !ok {
			return false, false
		}
		if !hasRows {
			return true, false
		}
	}

	return true, true
}

func (c *Migration) Run(driver database.Driver) {
	ok, hasRows := c.RunCheck(driver)
	if !ok && (len(c.OnFail) > 0) {
		err := driver.ExecuteBatch(c.OnFail)
		helpers.CheckError(err)
	}
	if ok && (len(c.OnSuccess) > 0) {
		err := driver.ExecuteBatch(c.OnSuccess)
		helpers.CheckError(err)
	}

	if ok && !hasRows && (len(c.OnNoResults) > 0) {
		err := driver.ExecuteBatch(c.OnNoResults)
		helpers.CheckError(err)
	}
	if ok && hasRows && (len(c.OnResults) > 0) {
		err := driver.ExecuteBatch(c.OnResults)
		helpers.CheckError(err)
	}
}

func (c *Migration) ResolveFiles(dir string) {
	if c.CheckFile != "" {
		c.CheckFile = filepath.Join(dir, c.CheckFile)
		content, err := os.ReadFile(c.CheckFile)
		helpers.CheckError(err)
		c.Check = append(c.Check, string(content))
	}
	if c.OnFailFile != "" {
		c.OnFailFile = filepath.Join(dir, c.OnFailFile)
		content, err := os.ReadFile(c.OnFailFile)
		helpers.CheckError(err)
		c.OnFail = append(c.OnFail, string(content))
	}
	if c.OnSuccessFile != "" {
		c.OnSuccessFile = filepath.Join(dir, c.OnSuccessFile)
		content, err := os.ReadFile(c.OnSuccessFile)
		helpers.CheckError(err)
		c.OnSuccess = append(c.OnSuccess, string(content))
	}
	if c.OnNoResultsFile != "" {
		c.OnNoResultsFile = filepath.Join(dir, c.OnNoResultsFile)
		content, err := os.ReadFile(c.OnNoResultsFile)
		helpers.CheckError(err)
		c.OnNoResults = append(c.OnNoResults, string(content))
	}
	if c.OnResultsFile != "" {
		c.OnResultsFile = filepath.Join(dir, c.OnResultsFile)
		content, err := os.ReadFile(c.OnResultsFile)
		helpers.CheckError(err)
		c.OnResults = append(c.OnResults, string(content))
	}
}

func getValue(field string, unmarshal func(interface{}) error) []string {
	mstr := make(map[string]string)
	if _ = unmarshal(&mstr); len(mstr) != 0 {
		if str, ok := mstr[field]; ok {
			return []string{str}
		}
	}

	miface := make(map[interface{}]interface{})
	if err := unmarshal(&miface); err == nil {
		sstr := make([]string, 0)
		if val, ok := miface[field]; ok {
			for _, v := range val.([]interface{}) {
				if str, ok := v.(string); ok {
					if strings.HasSuffix(str, ";") {
						sstr = append(sstr, str)
					} else {
						sstr = append(sstr, str+";")
					}
				}
			}
			return sstr
		}
	}
	return []string{}
}
