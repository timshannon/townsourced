// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */


// I've moved message processing into a separate file so its heavy
//  download doesn't have to be paid multiple times across multiple
//  pages.  Hopefully I won't have to do this too much, but if I do
//  I'll continue to use the ts global namespace

import {
    processMessage,
	escape,
}
from "./ts/message";

import emoji from "./ts/emoji";
import "./lib/emojione";

(function(ns) {
    "use strict";
    ns.processMessage = processMessage;
    ns.escapeMessage = escape;
    ns.emojiCategory = emoji.category;
    ns.emojiSearch = emoji.search;
    ns.emojiList = emojione.emojioneList;

}(window.ts = window.ts || {}));
