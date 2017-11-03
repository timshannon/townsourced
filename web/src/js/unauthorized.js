// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive + components
import Ractive from "./ts/ractiveInit";
import Login from "./components/login";

// ts libs
import {
    err
}
from "./ts/error";

//3rd party


$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "#main",
        template: "#tMain",
        components: {
			login: Login,
        },
		oncomplete: function() {
                    $("#username").focus();
		},
    });



});
