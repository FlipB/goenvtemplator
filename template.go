package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	//	"io/ioutil"
	"bytes"
	"io/ioutil"
	"log"
	"strings"
)

var alwaysThere string = "always_there"

func Env(key string) string {
	// We need to simulate always existing key for testing, but we do not want to pollute env
	if key == "ALWAYS_THERE" {
		return alwaysThere
	}

	value, ok := os.LookupEnv(key)
	if !ok {
		return ""
	}

	return value
}

func EnvList(key string) []string {
	// We need to simulate always existing key for testing, but we do not want to pollute env
	if key == "ALWAYS_THERE_LIST" {
		return []string{alwaysThere, alwaysThere}
	}

	value, ok := os.LookupEnv(key)
	if !ok {
		return []string{}
	}

	values := strings.Split(value, ",")
	if len(values) == 0 {
		return []string{}
	}

	optionalValues := make([]string, 0, len(values))
	for i := range values {
		optionalValues = append(optionalValues, values[i])
	}

	return optionalValues
}

func Default(args ...interface{}) (string, error) {
	for _, arg := range args {
		if arg == nil {
			continue
		}
		switch v := arg.(type) {
		case string:
			if arg.(string) != "" {
				return v, nil
			}
		case *string:
			if v != nil {
				return *v, nil
			}
		default:
			return "", fmt.Errorf("Default: unsupported type '%T'!", v)
		}
	}

	return "", errors.New("Default: all arguments are nil!")
}

func Require(arg interface{}) (string, error) {
	if arg == nil {
		return "", errors.New("Required argument is missing!")
	}

	switch v := arg.(type) {
	case string:
		return v, nil
	case *string:
		if v != nil {
			return *v, nil
		}
	}

	return "", fmt.Errorf("Requires: unsupported type '%T'!", v)
}

var funcMap = template.FuncMap{
	"env":     Env,
	"default": Default,
	"require": Require,
	"envlist": EnvList,
}

func generateTemplate(source, name string) (string, error) {
	var t *template.Template
	var err error
	if t, err = template.New(name).Funcs(funcMap).Parse(source); err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err = t.Execute(&buffer, nil); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func generateFile(templatePath, destinationPath string, debugTemplates bool) error {
	/*
		if !filepath.IsAbs(templatePath) {
			return fmt.Errorf("Template path '%s' is not absolute!", templatePath)
		}

		if !filepath.IsAbs(destinationPath) {
			return fmt.Errorf("Destination path '%s' is not absolute!", destinationPath)
		}
	*/

	var slice []byte
	var err error
	if slice, err = ioutil.ReadFile(templatePath); err != nil {
		return err
	}
	s := string(slice)
	result, err := generateTemplate(s, filepath.Base(templatePath))
	if err != nil {
		return err
	}

	if debugTemplates {
		log.Printf("Printing parsed template to stdout. (It's delimited by 2 character sequence of '\\x00\\n'.)\n%s\x00\n", result)
	}

	if err = ioutil.WriteFile(destinationPath, []byte(result), 0664); err != nil {
		return err
	}

	return nil
}
