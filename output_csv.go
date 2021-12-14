package main

import (
	"fmt"
	"os"
	"sort"
	//"strconv"
	"sync"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// CSVOutput writes the results of license lookups to an CSV file.
type CSVOutput struct {
	// Path is the path to the file to write. This will be overwritten if
	// it exists.
	Path string

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	modules map[*module.Module]interface{}
	lock    sync.Mutex
}

// Start implements Output
func (o *CSVOutput) Start(m *module.Module) {}

// Update implements Output
func (o *CSVOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *CSVOutput) Finish(m *module.Module, l *license.License, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.modules == nil {
		o.modules = make(map[*module.Module]interface{})
	}

	o.modules[m] = l
	if err != nil {
		o.modules[m] = err
	}
}

// Close implements Output
func (o *CSVOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	// Headers
	// Name,Version,License Type,Publisher,Description,URL,Comment

	// Sort the modules by name
	keys := make([]string, 0, len(o.modules))
	index := map[string]*module.Module{}
	for m := range o.modules {
		keys = append(keys, m.Path)
		index[m.Path] = m
	}
	sort.Strings(keys)

	file, err := os.OpenFile(o.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("File '%s' does not exists or cannot be created\n", o.Path)
		os.Exit(1)
	}
	defer file.Close()

	// Go through each module and output it into the spreadsheet
	for _, k := range keys {
		m := index[k]

		fmt.Fprintf(file, "%s,", m.Path)
		fmt.Fprintf(file, "%s,", m.Version)

		raw := o.modules[m]
		if raw == nil {
			fmt.Fprintf(file, "no, no")
			fmt.Fprintf(file, "\n")
			continue
		}

		// If the value is an error, then note the error
		if err, ok := raw.(error); ok {
			fmt.Fprintf(file, "ERROR: %s", err)
			fmt.Fprintf(file, "\n")
			continue
		}

		// If the value is a license, then mark the license
		if lic, ok := raw.(*license.License); ok {
			if lic != nil {
				fmt.Fprintf(file, "%s,", lic.SPDX)
			}

			fmt.Fprintf(file, ",")                // Publisher
			fmt.Fprintf(file, ",")                // Description
			fmt.Fprintf(file, "%s", lic.String()) // Comment
		}

		fmt.Fprintf(file, "\n")
	}

	return nil
}
