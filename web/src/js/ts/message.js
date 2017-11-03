// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true*/
import "../lib/commonmark";
import emoji from "./emoji";
import * as storage from "./storage";

//commonmark settings
var cmRead = new commonmark.Parser();
var cmWrite = new commonmark.HtmlRenderer({
    sourcepos: true,
    safe: true,
    smart: true,
});

cmWrite.softbreak = "<br />";

var rxUsername = /\B\\?@{1}[a-zA-Z0-9-]+/g;
var rxTown = /\s\\?\/town\/[a-zA-Z0-9-]+/g;
var rxHash = /\B\\?#{1}[a-zA-Z0-9-_]+/g;
var rxPrice = /\B\\?\$[0-9](,*[0-9]*)*(\.*[0-9]){0,2}/g;

export

function processMessage(msg) {
    "use strict";
    msg = msg || "";
    msg = emoji.toMarkdown(msg.replace(rxUsername, username)
        .replace(rxTown, townMatch)
        .replace(rxHash, hash)
        .replace(rxPrice, price));
    //replace #hashtag with a root level hash tag search for that tag /search?tag=hashtag
    //	:emoji-shortcuts:
    // then run it all through the commonmark parser
    return cmWrite.render(cmRead.parse(msg));
}


function username(match) {
    "use strict";
    match = match.trim();
    if (match[0] == "\\") {
        //escaped
        return match.slice(1);
    }
    var user = match.slice(1);
    storage.addUser(user);
    return "[" + match + "](/user/" + user + ")";
}


function townMatch(match) {
    "use strict";
    match = match.trim();
    if (match[0] == "\\") {
        //escaped
        return match.slice(1);
    }

    storage.addTown(match.split("/").pop());
    return "[" + match + "](" + match + ")";
}

function hash(match) {
    "use strict";
    match = match.trim();
    if (match[0] == "\\") {
        //escaped
        return match.slice(1);
    }

    // if in preset list of hashes, use hash tag image
    var replace;
    match = match.slice(1).toLowerCase();
    var tag = emoji.tag(match);
    if (tag) {
        replace = '![' + tag.name + '](/images/tags/' + tag.tag + '.png "tagged - ' + tag.name + '")';
    } else {
        replace = "#" + match;
    }

    return "[" + replace + "](/search?tag=" + match + ")";
}

function price(match) {
    "use strict";
    match = match.trim();
    if (match[0] == "\\") {
        //escaped
        return match.slice(1);
    }

    return '[' + match + '](# "price - ' + match + '")';
}



export

function escape(msg) {
    // escapes all instances of townsourced specific message items, like price, hashtags, mentions, etc
    "use strict";

    var escFunc = function(match) {
        match = match.trim();
        if (match[0] == "\\") {
            //escaped
            return match;
        }
        return "\\" + match;
    };

    msg = msg || "";

    msg = msg.replace(rxTown, escFunc)
        .replace(rxHash, escFunc)
        .replace(rxPrice, escFunc);

    return msg;
}
