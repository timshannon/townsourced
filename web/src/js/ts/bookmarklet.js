// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.

/* jshint  strict: true, esnext: true*/

import {
    origin,
    endsWith,
}
from "./util";


export
var bookmarkletElementID = "tsBookmarklet";

export

function bookmarklet() {
    "use strict";

    return encodeURIComponent("(function(){window.baseUrl='" + origin() + "';" +
        "var n = document.getElementById('" + bookmarkletElementID + "');" +
        "if(!n) {" +
        "var s=document.createElement('script');" +
        "s.setAttribute('id', '" + bookmarkletElementID + "'); " +
        "s.setAttribute('type','text/javascript');" +
        "s.setAttribute('charset','UTF-8');" +
        "s.setAttribute('src',baseUrl+'/js/bookmarklet.js');" +
        "document.documentElement.appendChild(s);}})()");

}

export

function autoParse() {
    "use strict";

    var domain = origin().toLowerCase();

    for (var i = 0; i < parsers.length; i++) {
        if (parsers[i].match(domain)) {
            return parsers[i].parse();
        }
    }

    return null;
}

/* parseResult = {
	title: "",
	content: "",
	images: [],
  }

  If parser doesn't find what it's looking for, it should return null, so that the manual selector can be used

*/

var parsers = [
    serverParser("craigslist.org"),
    serverParser("ebay.com"),
    serverParser("kijiji.ca"),
    serverParser("ebayclassifieds.com"), {
        match: function(str) {
            "use strict";
            return endsWith(str, "offerupnow.com");
        },
        parse: function() {
            "use strict";
            var result = {
                title: "",
                content: "",
                images: [],
            };

            // title
            var titleEl = document.body.querySelector("h1.title");
            if (!titleEl) {
                return null;
            }
            result.title = titleEl.textContent;

            // content

            var contentEl = document.body.querySelector("div[itemprop=description]");
            if (!contentEl) {
                return null;
            }
            result.content = contentEl.innerHTML;

            var conditionEl = contentEl.nextElementSibling;
            if (conditionEl) {
                if (conditionEl.nextElementSibling) {
                    result.content += "<br><br><strong>Condition:</strong> " + conditionEl.nextElementSibling.textContent;
                }
            }

            // images
            var gallery = document.getElementById("carousel-gallery-container");
            if (!gallery) {
                return null;
            }

            var images = gallery.getElementsByTagName("img");

            for (var i = 0; i < images.length; i++) {
                result.images.push(images[i].getAttribute("src"));
            }

            return result;
        },
    },
];

function serverParser(domain) {
    "use strict";
    return {
        match: function(str) {
            return endsWith(str, domain);
        },
        parse: function() {
            return {};
        },
    };

}
