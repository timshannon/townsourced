// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */

import "../lib/emojione";
import {
    unique,
    find,
}
from "./util";
import data from "./emojiData";

var tags = [{
    tag: "pricereduced",
    name: "price reduced",
    keywords: ["sale", "bargain", "must sell", "reduced", "slashed", "cheap", "must go"],
}, {
    tag: "CurbAlert",
    name: "curb alert",
    keywords: ["free", "freebies", "must pick up", "give away", "porch pick up"],
}, {
    tag: "CrossPosted",
    name: "cross posted",
    keywords: ["other site", "multiple posts"],
}, {
    tag: "freebies",
    name: "freebies",
    keywords: ["curb alert", "free", "must go", "give away"],
}, {
    tag: "PM",
    name: "private message",
    keywords: ["email", "private", "personal"],
}, {
    tag: "OBO",
    name: "or best offer",
    keywords: ["make an offer", "auction", "must sell", "cheap", "must go"],
}, {
    tag: "NWT",
    name: "new with tags",
    keywords: ["original"],
}, {
    tag: "NWOT",
    name: "new with out tags"
}, {
    tag: "SF",
    name: "smoke free",
    keywords: ["smell", "clean", "odor", "allergy"],
}, {
    tag: "PF",
    name: "pet free",
    keywords: ["smell", "clean", "odor", "allergy"],
}, {
    tag: "SFPF",
    name: "smoke free pet free",
    keywords: ["smell", "clean", "odor", "allergy"],
}, {
    tag: "MPU",
    name: "must pick up",
    keywords: ["delivery"],
}, {
    tag: "PPU",
    name: "pending pick up",
    keywords: ["waiting", "limbo"],
}, {
    tag: "ISO",
    name: "in search of",
    keywords: ["looking", "seeking", "scouting", "wanted"],

}, {
    tag: "EUC",
    name: "excellent used condition",
    keywords: ["quality", "status", "shape", "form"],
}, {
    tag: "GUC",
    name: "good used condition",
    keywords: ["quality", "status", "shape", "form"],
}, {
    tag: "FUC",
    name: "fair used condition",
    keywords: ["quality", "status", "shape", "form"],
}, {
    tag: "ORP",
    name: "original retail price",
    keywords: ["same price", "list price", "value", "store price"],
}, {
    tag: "TIA",
    name: "thanks in advance",
    keywords: ["thankful", "appreciate", "kind"],
}, {
    tag: "NIB",
    name: "new in box",
    keywords: ["wrapper"],
}, {
    tag: "NIP",
    name: "new in package",
    keywords: ["wrapper"],
}, {
    tag: "DERP",
    name: "don't engage report post (to moderator)",
    keywords: ["moderator"],
}, {
    tag: "HTF",
    name: "hard to find",
    keywords: ["rare", "limited", "uncommon", "unique", "unusual"],
}, {
    tag: "HM",
    name: "hand made",
    keywords: ["home made", "craft", "craftmanship", "unique"],
}, ];

data.category.tags = tags;
data.keyword = data.keyword.concat(tags);

var maxSearchResult = 50;

export
default {
    toMarkdown: function(msg) {
        "use strict";
        return emojione.toMarkdown(msg);
    },

    // returns the list of emoji that match the category
    category: function(cat) {
        "use strict";
        return data.category[cat];
    },
    tag: function(t) {
        "use strict";

        return tags.filter(function(tag) {
            return tag.tag.toLowerCase() === t.toLowerCase();
        })[0];
    },

    // returns a list of emoji that match the src string
    search: function(str) {
        "use strict";
        var results = [];
        str = str.toLowerCase();

        var filters = [

            function(e) {
                //name starts with string
                return e.name.indexOf(str) === 0;
            },
            function(e) {
                if (!e.keywords) {
                    return false;
                }
                //has a keyword that starts with string
                return e.keywords.filter(function(k) {
                    return k.indexOf(str) === 0;
                }).length > 0;
            },
            function(e) {
                //name contains string
                return e.name.indexOf(str) !== -1;
            },
            function(e) {
                //has a keword that contains string
                if (!e.keywords) {
                    return false;
                }
                return e.keywords.filter(function(k) {
                    return k.indexOf(str) !== -1;
                }).length > 0;
            },

        ];

        for (var i = 0; i < filters.length; i++) {
            results = results.concat(data.keyword.filter(filters[i]));
            results = unique(results);
            if (results.length >= maxSearchResult) {
                break;
            }
        }

        return results.slice(0, maxSearchResult);
    },
};



//emojione settings
emojione.imageType = "png";
emojione.imagePathPNG = "/images/emoji/png/";
emojione.ascii = true;

(function(ns) {
    "use strict";
    // emojione currently doesn't have a full conversion of ascii and unicode to image path
    // they instead return the full dom object.  I don't want that, so I'm copying most
    // of their code.
    ns.toMarkdown = function(str) {
        str = ns.unicodeToMarkdown(str);
        str = ns.shortnameToMarkdown(str);
        return str;
    };


    ns.unicodeToMarkdown = function(str) {
        var replaceWith, unicode, alt;

        if ((!ns.unicodeAlt) || (ns.sprites)) {
            // if we are using the shortname as the alt tag then we need a reversed array to map unicode code point to shortnames
            var mappedUnicode = ns.mapShortToUnicode();
        }

        str = str.replace(new RegExp("<object[^>]*>.*?<\/object>|<span[^>]*>.*?<\/span>|<(?:object|embed|svg|img|div|span|p|a)[^>]*>|(" + 
		ns.unicodeRegexp + ")", "gi"),            function(unicodeChar) {
                if ((typeof unicodeChar === 'undefined') || (unicodeChar === '') || (!(unicodeChar in ns.jsEscapeMap))) {
                    // if the unicodeChar doesnt exist just return the entire match
                    return unicodeChar;
                } else {
                    // get the unicode codepoint from the actual char
                    unicode = ns.jsEscapeMap[unicodeChar];

                    // depending on the settings, we'll either add the native unicode as the alt tag, otherwise the shortname
                    alt = (ns.unicodeAlt) ? ns.convert(unicode.toUpperCase()) : mappedUnicode[unicode];

                    // depending on the settings, we'll either add the native unicode as the alt tag, otherwise the shortname
                    alt = (ns.unicodeAlt) ? ns.convert(unicode) : mappedUnicode[unicode];

                    if (ns.imageType === 'png') {
                        replaceWith = "![" + alt + "](" + ns.imagePathPNG + unicode + ".png" + ns.cacheBustParam + " 'emoji - " + alt + " ')";
                    } else {
                        // svg
                        replaceWith = "![" + alt + "](" + ns.imagePathSVG + unicode + ".svg" + ns.cacheBustParam + " 'emoji - " + alt + "')";
                    }



                    return replaceWith;
                }
            });

        return str;
    };
    ns.shortnameToMarkdown = function(str) {

        // replace regular shortnames first
        var replaceWith, unicode, alt;
        str = str.replace(new RegExp("<object[^>]*>.*?<\/object>|<span[^>]*>.*?<\/span>|<(?:object|embed|svg|img|div|span|p|a)[^>]*>|(" +
            ns.shortnames + ")", "gi"), function(shortname) {
            if ((typeof shortname === 'undefined') || (shortname === '') || (!(shortname in ns.emojioneList))) {
                // if the shortname doesnt exist just return the entire match
                return shortname;
            } else {
                unicode = ns.emojioneList[shortname][ns.emojioneList[shortname].length - 1];

                // depending on the settings, we'll either add the native unicode as the alt tag, otherwise the shortname
                alt = (ns.unicodeAlt) ? ns.convert(unicode.toUpperCase()) : shortname;
                if (ns.imageType === 'png') {
                    replaceWith = "![" + alt + "](" + ns.imagePathPNG + unicode + ".png" + ns.cacheBustParam + " 'emoji - " + shortname + "')";
                } else {
                    // svg
                    replaceWith = "![" + alt + "](" + ns.imagePathSVG + unicode + ".svg" + ns.cacheBustParam + " 'emoji - " + shortname + "')";
                }

                return replaceWith;
            }
        });

        // if ascii smileys are turned on, then we'll replace them!
        if (ns.ascii) {

            str = str.replace(new RegExp("<object[^>]*>.*?<\/object>|<span[^>]*>.*?<\/span>|<(?:object|embed|svg|img|div|span|p|a)[^>]*>|((\\s|^)" +
                ns.asciiRegexp + "(?=\\s|$|[!,.?]))", "g"), function(entire, m1, m2, m3) {
                if ((typeof m3 === 'undefined') || (m3 === '') || (!(ns.unescapeHTML(m3) in ns.asciiList))) {
                    // if the shortname doesnt exist just return the entire match
                    return entire;
                }

                m3 = ns.unescapeHTML(m3);
                unicode = ns.asciiList[m3];

                // depending on the settings, we'll either add the native unicode as the alt tag, otherwise the shortname
                alt = (ns.unicodeAlt) ? ns.convert(unicode.toUpperCase()) : ns.escapeHTML(m3);
                if (ns.imageType === 'png') {
                    replaceWith = "![" + alt + "](" + ns.imagePathPNG + unicode + ".png" + ns.cacheBustParam + " 'emoji - " + ns.escapeHTML(m3) + "')";
                } else {
                    // svg
                    replaceWith = "![" + alt + "](" + ns.imagePathSVG + unicode + ".svg" + ns.cacheBustParam + " 'emoji - " + ns.escapeHTML(m3) + "')";
                }



                return replaceWith;
            });
        }

        return str;
    };
}(window.emojione = window.emojione || {}));
