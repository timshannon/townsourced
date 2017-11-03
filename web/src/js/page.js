// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
//
//components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";

import {
    htmlPayload,
}
from "./ts/util";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
        },
        data: function() {
            return htmlPayload();
        },
    });

});
