Web Source
===============

Plan is as follows:

* Build with gobble
*	CSS
	* One compiled townsourced.min.css file built with less from bootstrap and fontawesome source
* JS 
	* Built with esperanto + amdclean + gobble
	* Ractive Components built to es6 via gobble-ractive, then everything is run through esperanto
	* jquery & ractive loaded globally
		* jquery loaded from google cdn
		* ractive loaded locally
	* ts.js - API library for interacting with the townsourced API, and any other global utility functions
		* may end up being loaded as modules instead of one global library
* Minimum Browser support is IE 9+
* There will be one project level gobble file for building all CSS and JS


##Ractive Usage
In order to keep the UI as responsive as possible, all concrete page elements should be defined in Vanilla HTML so they are presented immediatly without flicker.  If that is not possible, then a placeholder HTML element should be created that will be replaced by the Ractive defined elements when they finish loading.  This will prevent flickerinng and resizing of page elements on initial load.

Ractives should be as self contained as possible.  Stay away from large global ractive elements, and make smaller, well defined ractive elements that are responsible for their own areas.  This can be reconsidered if keeping two ractives separate creates a lot of complexity for sharing data.
