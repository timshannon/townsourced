// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

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
			return {
				help: htmlPayload(),
				processed: null,	
			};
		},
		oncomplete: function() {
			this.set("processed", ts.processMessage(this.get("help.document")));
		},
    });

});
