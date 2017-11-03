//Copyright (c) 2015 Tim Shannon
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in
//all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
//THE SOFTWARE.

package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

//Cfg is a container for reading and writing to a simple
// JSON config file, nothing fancy.  Easy to parse,
// easy to read and edit by humans
type Cfg struct {
	fileName       string
	values         map[string]interface{}
	autoWrite      bool
	isEnv          bool
	variablePrefix string
}

//Load loads the config file from the passed in config path
// If the config file cannot be parsed (i.e. isn't valid json) or
// cannot be found an error will be returned
// filename can be a slice of filenames, the first file found is loaded
// if none of the files in the slice are found, an error is returned
func Load(filename ...string) (*Cfg, error) {
	for i := range filename {
		c := &Cfg{fileName: filename[i]}
		err := c.Load()
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	return nil, os.ErrNotExist
}

//LoadOrCreate automatically creates the passed in config file if it
// doesn't already exist, then load it.  If file gets created, then all
// values loaded afterwards will be "defaulted" and will be written to this
// new file
// filename can be a slice of filenames, the first file found is loaded
// if none of the files in the slice are found, the first file in the slice is created
func LoadOrCreate(filename ...string) (*Cfg, error) {
	c, err := Load(filename...)

	if os.IsNotExist(err) {
		c = &Cfg{
			fileName:  filename[0],
			autoWrite: true,
			values:    make(map[string]interface{}),
		}
		err = os.MkdirAll(filepath.Dir(filename[0]), 0777)
		if err != nil {
			return nil, errWithFile(c.FileName(), err)
		}
		err = c.Write()
	}
	if err != nil {
		return nil, err
	}

	return c, nil
}

//LoadEnv loads the CFG values from environment variables instead
// of from a file.  Environment variables will be prefixed with the
// the passed in variablePrefix
func LoadEnv(variablePrefix string) *Cfg {
	return &Cfg{
		isEnv:          true,
		variablePrefix: variablePrefix,
	}
}

//Load loads a config file from the passed in location
func (c *Cfg) Load() error {
	if c.isEnv {
		return nil
	}

	if c.fileName == "" {
		err := errors.New("No Filename set for Cfg object")
		return err
	}

	data, err := ioutil.ReadFile(c.fileName)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &c.values); err != nil {
		return errWithFile(c.FileName(), err)
	}
	return nil
}

// FileName is the path to the configuration file
func (c *Cfg) FileName() string {
	return c.fileName
}

//Value returns the raw interface value for the config entry with the given name
//  The type is left up to the consumer to determine
// if cfg is an ENV type, result will always be a string
func (c *Cfg) Value(name string, defaultValue interface{}) interface{} {
	if c.isEnv {
		value := os.Getenv(c.variablePrefix + name)
		if value == "" {
			return defaultValue
		}
		return value
	}
	value, ok := c.values[name]
	if !ok {
		if c.autoWrite {
			c.SetValue(name, defaultValue)
			c.Write()
		}
		return defaultValue
	}

	return value
}

//ValueToType allows you to pass in a struct as the result
// for which you want to load the config entry into
// Marshalls the JSON data directly into your passed in type
// If the value doesn't exist, then the passed in result value
// will be set as the default
func (c *Cfg) ValueToType(name string, result interface{}) error {
	if c.isEnv {
		value := os.Getenv(c.variablePrefix + name)
		if value == "" {
			return nil
		}

		err := json.Unmarshal([]byte(value), result)
		return errWithFile(c.FileName(), err)
	}
	value, ok := c.values[name]
	if !ok {
		if c.autoWrite {
			c.SetValue(name, result)
			c.Write()
		}
		return nil
	}

	//marshall value
	j, err := json.Marshal(value)
	if err != nil {
		return errWithFile(c.FileName(), err)
	}

	err = json.Unmarshal(j, result)
	if err != nil {
		return errWithFile(c.FileName(), err)
	}
	return nil
}

//Int retrieves an integer config value with the given name
// if a value with the given name is not found the default is returned
func (c *Cfg) Int(name string, defaultValue int) int {
	if c.isEnv {
		value := os.Getenv(c.variablePrefix + name)
		if value == "" {
			return defaultValue
		}
		i, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return i
	}
	value, ok := c.values[name].(float64)
	if !ok {
		if c.autoWrite {
			c.SetValue(name, defaultValue)
			c.Write()
		}
		return defaultValue
	}
	return int(value)
}

//String retrieves a string config value with the given name
// if a value with the given name is not found the default is returned
func (c *Cfg) String(name string, defaultValue string) string {
	if c.isEnv {
		value := os.Getenv(c.variablePrefix + name)
		if value == "" {
			return defaultValue
		}
		return value
	}
	value, ok := c.values[name].(string)
	if !ok {
		if c.autoWrite {
			c.SetValue(name, defaultValue)
			c.Write()
		}
		return defaultValue
	}
	return value
}

//Bool retrieves a bool config value with the given name
// if a value with the given name is not found the default is returned
func (c *Cfg) Bool(name string, defaultValue bool) bool {
	if c.isEnv {
		value := os.Getenv(c.variablePrefix + name)
		if value == "" {
			return defaultValue
		}
		b, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return b
	}
	value, ok := c.values[name].(bool)
	if !ok {
		if c.autoWrite {
			c.SetValue(name, defaultValue)
			c.Write()
		}
		return defaultValue
	}
	return value
}

//Float retrieves a float config value with the given name
// if a value with the given name is not found the default is returned
func (c *Cfg) Float(name string, defaultValue float32) float32 {
	if c.isEnv {
		value := os.Getenv(c.variablePrefix + name)
		if value == "" {
			return defaultValue
		}
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return defaultValue
		}
		return float32(f)
	}
	value, ok := c.values[name].(float64)

	if !ok {
		if c.autoWrite {
			c.SetValue(name, defaultValue)
			c.Write()
		}

		return defaultValue
	}
	return float32(value)
}

//SetValue sets a config value.  It is left up to the end user
// to then write out the new values with the .Write() function
// Environment variables aren't set to any value.
func (c *Cfg) SetValue(name string, value interface{}) {
	if c.isEnv {
		return
	}

	c.values[name] = value
}

//Write writes the config values to the config's file location
// for Env CFG's nothing is written, as the env variables are
// set immediately
func (c *Cfg) Write() error {
	if c.isEnv {
		return nil
	}
	if c.fileName == "" {
		return errors.New("No FileName set for this config")
	}
	data, err := json.MarshalIndent(c.values, "", "    ")
	if err != nil {
		return errWithFile(c.FileName(), err)
	}
	err = ioutil.WriteFile(c.fileName, data, 0666)

	return err

}

//errWithFile returns the passed in error with the context of which selected filename
// the error occurred on
func errWithFile(filename string, err error) error {
	return errors.New("Error processing config file " + filename + ": " + err.Error())
}
