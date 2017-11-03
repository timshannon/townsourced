Config
=====

Simple JSON config file.  

To Use
------

	import "bitbucket.org/tshannon/config"
	
	//Loads the json file, creates it if one doesn't exist
	cfg, err := LoadOrCreate("settings.json")
	if err != nil {
		panic(fmt.Sprintf("Cannot load settings.json: %v", err)
	}

	//Return the value of url, if not found use the default "http://google.com"
	cfg.String("url", "http://bing.com")

	cfg.Set("url", "http://google.com")
	err := cfg.Write()
	if err != nil {
		panic("Cannot write settings.json: %v", err)
	}

You can pass a slice filenames into the Load and LoadOrCreate, and it will use the first one it finds.  If none are found when passed into LoadOrCreate, then the first file in the slice will be created.

You can also use the ```StandardFileLocations``` which will return a slice of standard config file locations for your OS.  Generally follows this priority list:

 1. User locations are used before...
 2. System locations which are used before ...
 3. The imediate running directory of the application

 The result set will be joined with the passed in filepath.  Passing in a filepath with a leading directory is encouraged to keep your config folders organized.

 For example a filepath of myApp/config.json might return the following on linux:
```
 	"/home/user/.config/myApp/config.json",
	"/etc/xdg/myApp/config.json",
	"/etc/myApp/config.json",
	"./config.json"
```

 Note that parent folder paths (myApp in this example) are stripped for the last eligble file location so the config file will exist in the same directory as the running executable.


Below is a complete example of an application which would support configuration being supplied either via config files in standard OS locations,  or Environment variables.
```
package main

import (
	"fmt"
	"os"

	"bitbucket.org/tshannon/config"
)

func main() {

	cfg, err := config.Load(config.StandardFileLocations("myApp/config.json")...)

	if os.IsNotExist(err) {
		// On Linux, no config files found in:
		//	/home/user/.config/myApp/config.json
		//	/etc/xdg/myApp/config.json
		//	/etc/myApp/config.json
		//	./config.json
		cfg = config.LoadEnv("MYAPP_")
	}

	address := cfg.String("address", "127.0.0.1")
	fmt.Println(address)
	//should print 127.0.0.1 on a fresh machine
	//export MYAPP_address=192.168.1.1 then run again to print 192.168.1.1

}

```
