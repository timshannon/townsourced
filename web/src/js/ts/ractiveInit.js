// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// current just disabling DEBUG mode on minification, but it could init other global Ractive settings as well

import Ractive from "ractive";

Ractive.DEBUG = /unminified/.test(function() { /*unminified*/ });

export
default Ractive;
