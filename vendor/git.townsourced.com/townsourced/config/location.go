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
	"os"
	"path/filepath"
)

//StandardFileLocations builds an OS specific list of standard file locations
// for where a config file should be loaded from.
// Generally follows this priority list:
// 1. User locations are used before...
// 2. System locations which are used before ...
// 3. The imediate running directory of the application
// The result set will be joined with the passed in filepath.  Passing in
// a filepath with a leading directory is encouraged to keep your config folders
// clean.
//
// For example a filepath of myApp/config.json might return the following on linux
// 	"/home/user/.config/myApp/config.json",
//	"/etc/xdg/myApp/config.json",
//	"./config.json"
// Note that parent folder paths (myApp in this example) are stripped for the first eligible file location
// so the config file will exist in the same directory as the running executable
func StandardFileLocations(cfgPath string) []string {
	locations := append(userLocations(), systemLocations()...)

	for i := range locations {
		if locations[i] != "" {
			locations[i] = filepath.Join(locations[i], cfgPath)
		}
	}

	//get running dir
	runningDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		runningDir = "."
	}
	runningDir = filepath.Join(runningDir, filepath.Base(cfgPath))

	locations = append(locations, runningDir)
	return locations
}
