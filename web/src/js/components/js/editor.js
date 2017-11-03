// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

import "../../lib/jquery.atwho";
import "../../lib/jquery.caret";

import {
    get as storageGet
}
from "../../ts/storage";

import {
    unique
}
from "../../ts/util";

var timer;

export
default {
    data: {
        input: "",
        parsed: "",
        linkURL: "",
        ecat: "people",
        emojiSearch: "",
    },
    decorators: {
        editor: function(node) {
            "use strict";
            var r = this;

            r.set("editNode", node);

            autocomplete(node, r);
            return {
                teardown: function() {
                    r.set("editNode", null);
                    autocomplete(node, r, true);
                }
            };
        },
    },

    onrender: function() {
        "use strict";
        var r = this;

        if (!r.get("input")) {
            r.set("input", "");
        }

        atData();

        r.on({
            "link": function(event, link) {
                event.original.preventDefault();
                window.open(link);
            },
            "setSelection": function(start, end) {
                var editor = r.get("editNode");
                if (start !== undefined && end === undefined) {
                    end = start;
                }
                if (end !== undefined && start === undefined) {
                    start = end;
                }

                if (editor.setSelectionRange) {
                    editor.focus();
                    editor.setSelectionRange(start, end);
                }
            },
            "gotoSource": function(event) {
                //capture ctrl-click
                // TODO: Not currently used as source-pos isn't that accurate
                if (event.original.ctrlKey) {
                    event.original.preventDefault();
                    var s = sourcePosToSelection(getSourcePos(event.original.target));

                    r.fire("setSelection", s[0], s[1]);
                }
            },
            "keydown": function(event) {
                var e = event.original;

                // tab key handling
                if (e.keyCode === 9 && !e.ctrlKey && !e.altKey && !e.metaKey) {
                    var selection = [r.get("editNode").selectionStart, r.get("editNode").selectionEnd];
                    if (selection[0] !== selection[1]) {
                        e.preventDefault();
                        if (e.shiftKey) {
                            r.fire("toolbar", null, "shifttab");
                        } else {
                            r.fire("toolbar", null, "tab");
                        }
                    } else if (!e.shiftKey) {
                        //if is a cursor and shift isn't down
                        //insert tab
                        e.preventDefault();

                        replaceAt("\t", selection);
                        r.fire("setSelection", selection[1] + 1);
                    }
                }

                // editor hotkeys

                // ctrl hotkeys
                if (e.ctrlKey && !e.altKey && !e.metaKey && !e.shiftKey) {
                    switch (e.keyCode) {
                        case 66: //<ctrl-b>
                            e.preventDefault();
                            r.fire("toolbar", null, "bold");
                            break;
                        case 73: //<ctrl-i>
                            e.preventDefault();
                            r.fire("toolbar", null, "italic");
                            break;
                    }

                }
            },
            "linkDropDown": function(event) {
                $("#links").on("shown.bs.dropdown", function() {
                    r.set("linkURL", "http://");

                    $("#linkURL").focus();
                    $("#linkURL")[0].setSelectionRange(7, 7);
                });
            },
            "loadEmoji": function(event) {
                //makes sure emoji are loaded
                if (r.get("ecat") !== "search") {
                    r.set("emojiList", ts.emojiCategory(r.get("ecat")));
                }
            },
            "emojiCat": function(event, category) {
                r.set("emojiList", []);
                event.original.stopPropagation();
                event.original.target.blur();
                event.original.target.parentNode.blur();
                r.set("emojiList", ts.emojiCategory(category));
                r.set("ecat", category);
            },
            "eSearch": function(event) {
                event.original.stopPropagation();
                event.original.target.blur();
                event.original.target.parentNode.blur();

                r.set("emojiList", null);
                r.set("ecat", "search");
                r.set("emojiSearch", "");
                $("#emojiSearch").focus();
            },
            "searchEmoji": function(event) {
                if (timer) {
                    window.clearTimeout(timer);
                }

                timer = window.setTimeout(function() {
                    r.set("emojiList", ts.emojiSearch(r.get("emojiSearch")));
                }, 100);
            },
            "selectEmoji": function(event) {
                var s = [r.get("editNode").selectionStart, r.get("editNode").selectionEnd];

                var change = replaceAt(event.context.shortname, s);

                r.fire("setSelection", s[1] + change);
            },
            "selectTag": function(event) {
                var s = [r.get("editNode").selectionStart, r.get("editNode").selectionEnd];

                var change = replaceAt(" #" + event.context.tag + " ", s);

                r.fire("setSelection", s[1] + change);

            },
            "toolbar": function(event, type) {
                r.set("selection", [r.get("editNode").selectionStart, r.get("editNode").selectionEnd]);
                var change = 0;
                switch (type) {
                    case "h1":
                        change = editLines(null, null, null, "# ", ["# ", "## ", "### "]);
                        break;
                    case "h2":
                        change = editLines(null, null, null, "## ", ["# ", "## ", "### "]);
                        break;
                    case "h3":
                        change = editLines(null, null, null, "### ", ["# ", "## ", "### "]);
                        break;
                    case "bold":
                        change = editWord("**");
                        break;
                    case "italic":
                        change = editWord("*");
                        break;
                    case "ul":
                        change = editLines(null, null, null, "* ", ["1. "]);
                        break;
                    case "ol":
                        change = editLines(null, null, null, "1. ", ["* "]);
                        break;
                    case "quote":
                        change = editLines("> ");
                        break;
                    case "link":
                        event.original.preventDefault();
                        change = addLink();
                        $("#mdLink").dropdown("toggle");
                        break;
                    case "unlink":
                        change = removeLinks();
                        break;
                    case "hr":
                        change = replaceAt("\n***\n");
                        break;
                    case "question":
                        break;
                    case "tab":
                        change = editLines(null, "\t");
                        break;
                    case "shifttab":
                        change = editLines(null, null, "\t");
                        break;

                }
                //Set selection at beginning and end of edit
                r.fire("setSelection", r.get("selection.0"), r.get("selection.1") + change);
            },
        });

        function replaceAt(str, selection) {
            selection = selection || r.get("selection");
            r.set("input", r.get("input").slice(0, selection[0]) + str + r.get("input").slice(selection[1]));
            return str.length - (selection[1] - selection[0]);
        }

        //in functions below, selection is [start, end]

        function editLines(toggle, add, remove, toggleOrReplace, replaceThese) {
            function addStr(line, str) {
                lines[line] = str + lines[line];
                return str.length;
            }

            function removeStr(line, str) {
                if (lines[line].indexOf(str) === 0) {
                    lines[line] = lines[line].slice(str.length);
                    return str.length;
                }
                return 0;
            }

            function toggleStr(line, str) {
                if (lines[line].indexOf(str) === 0) {
                    lines[line] = lines[line].slice(str.length);
                    return str.length;
                } else {
                    lines[line] = str + lines[line];
                    return str.length;
                }
                return 0;
            }

            function toggleOrReplaceStr(line, str) {
                if (lines[line].indexOf(str) === 0) {
                    return toggleStr(line, str);
                }

                var change = 0;

                for (var i = 0; i < replaceThese.length; i++) {
                    change += removeStr(line, replaceThese[i]);
                }

                change += addStr(line, str);

                return change;
            }

            var s = selectToLine();
            if (!s) {
                return 0;
            }
            r.set("selection", s);
            var lines = r.get("input").slice(s[0], s[1]).split("\n");
            var change = 0;

            for (var i = 0; i < lines.length; i++) {
                if (lines[i].trim() === "") {
                    continue;
                }
                if (toggle) {
                    change += toggleStr(i, toggle);
                }
                if (add) {
                    change += addStr(i, add);
                }
                if (remove) {
                    change += removeStr(i, remove);
                }
                if (toggleOrReplace) {
                    change += toggleOrReplaceStr(i, toggleOrReplace);
                }
            }

            var added = lines.join("\n");
            r.set("input", r.get("input").slice(0, s[0]) + added + r.get("input").slice(s[1]));
            return change;
        }


        function editWord(toggle, add, remove) {
            var s = selectToWord();
            if (!s) {
                return 0;
            }

            r.set("selection", s);
            var words = r.get("input").slice(s[0], s[1]).split("\n");
            var wrap = toggle || add || remove;

            for (var i = 0; i < words.length; i++) {
                var word = words[i];

                if (word.trim() === "") {
                    continue;
                }

                var wrapped = (word.indexOf(wrap) === 0 && word.slice(wrap.length * -1) === wrap);

                if (add || (toggle && !wrapped)) {
                    word = wrap + word + wrap;
                } else if (remove || toggle) {
                    if (wrapped) {
                        word = word.slice(wrap.length, word.length - wrap.length);
                    }
                }
                words[i] = word;
            }

            return replaceAt(words.join("\n"), s);
        }

        // extends the current selection to the nearest end of line
        function selectToLine() {
            return selectTo(["\n"]);
        }

        // update the current selection to the nearest word
        function selectToWord() {
            //FIXME: trim off leading spaces
            return selectTo([" ", "\t", "\n"]);
        }

        function selectTo(startChars, endChars) {
            var input = r.get("input");
            var selection = r.get("selection");
            endChars = endChars || startChars;

            if (selection[0] === selection[1]) {
                //is a cursor grow the selection one character in each direction
                // as long as it's not the selection ending char
                if (startChars.indexOf(input[selection[0] - 1]) === -1) {
                    selection[0]--;
                }
                if (endChars.indexOf(input[selection[1]]) === -1) {
                    selection[1]++;
                }
            }

            //if selection isn't a cursor, and empty space is selected, drop out early
            if (input.slice(selection[0], selection[1]).trim() === "") {
                return false;
            }

            for (var i = selection[0]; i >= 0; i--) {
                if (i === 0) {
                    selection[0] = i;
                    break;
                } else if (startChars.indexOf(input[i]) !== -1) {
                    selection[0] = i + 1;
                    break;
                }
            }

            //start end of selection back a single character to test if
            // end selection is already on an endchar
            selection[1]--;

            for (i = selection[1]; i <= input.length; i++) {
                if (endChars.indexOf(input[i]) !== -1) {
                    selection[1] = i;
                    break;
                } else if (i === input.length) {
                    selection[1] = i;
                    break;
                }
            }

            if (input.slice(selection[0], selection[1]).trim() === "") {
                return false;
            }


            return selection;
        }

        function getSourcePos(node) {
            var sourcepos = "";

            while (!$(node).is(".parsed")) {
                sourcepos = $(node).attr("data-sourcepos");
                if (sourcepos) {
                    return sourcepos;
                }
                node = node.parentNode;
            }

            return sourcepos;
        }

        // returns a selection from the commonmark sourcepos string
        function sourcePosToSelection(sourcepos) {
            var fRow, tRow, fCol, tCol; //from to row column

            var input = r.get("input");
            var s = [-1, -1];

            var ft = sourcepos.split("-");
            if (ft.length !== 2) {
                return [0, 0];
            }
            var rowCol = ft[0].split(":");
            if (rowCol.length !== 2) {
                return [0, 0];
            }


            fRow = Number(rowCol[0]);
            fCol = Number(rowCol[1]);

            rowCol = ft[1].split(":");
            if (rowCol.length !== 2) {
                return [0, 0];
            }
            tRow = Number(rowCol[0]);
            tCol = Number(rowCol[1]);

            var curRow = 1;
            var rowBegin = 0;
            //seek to row, sourcepos starts at 1
            for (var i = 0; i < input.length; i++) {
                if (input[i] == "\n") {
                    curRow++;
                    rowBegin = i + 1;
                }
                if (s[0] === -1) { //from not found yet
                    if (curRow === fRow) {
                        s[0] = rowBegin + (fCol - 1);
                    }
                } else { //from found look for to
                    if (curRow === tRow) {
                        s[1] = rowBegin + tCol;
                        break;
                    }
                }
            }


            if (s[0] === -1 || s[1] === -1) {
                return [0, 0];
            }


            return s;
        }


        function removeLinks() {
            function removeLink(str) {
                function linkText(str) {
                    str = str.slice(1);
                    for (var i = 0; i < str.length; i++) {
                        if (str[i] === "]") {
                            return str.slice(0, i);
                        }
                    }
                }
                var need = "[]()";
                var start = -1,
                    end = -1;


                for (var i = 0, n = 0; i < str.length; i++) {
                    if (str[i] === need[n]) {
                        if (n === 0) {
                            start = i;
                        }
                        if (n === (need.length - 1)) {
                            end = i;
                            break;
                        } else {
                            n++;
                        }
                    }
                }

                if (start === -1 || end === -1) {
                    return str;
                }

                return str.slice(0, start) + linkText(str.slice(start, end)) + str.slice(end + 1);
            }

            var s = selectToWord();
            if (!s) {
                return 0;
            }

            r.set("selection", s);
            var selected = "";
            var updated = r.get("input").slice(s[0], s[1]);
            while (selected !== updated) {
                selected = updated;
                updated = removeLink(selected);
            }

            return replaceAt(selected, s);

        }

        function addLink() {
            var s = selectToWord();

            var link = r.get("linkURL");
            if (!link || link.trim() === "http://") {
                return 0;
            }

            link = link.replace(/\s/g, "");

            r.set("selection", s);
            var selected = r.get("input").slice(s[0], s[1]) || link;

            selected = selected.replace(/\n/g, " ");

            selected = "[" + selected + "](" + link + ")";
            return replaceAt(selected, s);
        }



        function atData() {
            var atUsers = storageGet("atUsers") || [];
            var atTowns = storageGet("towns") || [];

            if (r.get("atUsers")) {
                atUsers = atUsers.concat(r.get("atUsers"));
            }

            if (r.get("atTowns")) {
                atTowns = atTowns.concat(r.get("atTowns"));
            }
            r.set("atUserData", unique(atUsers, "username"));
            r.set("atTownData", unique(atTowns));
        }


        r.observe({
            "atUsers": function(newVal) {
                if (!newVal) {
                    return;
                }

                r.set("atUserData", unique(r.get("atUserData").concat(newVal), "username"));
            },
            "atUserData": function(newVal) {
                if (!newVal) {
                    return;
                }
                var n = $(r.get("editNode"));
                if (n.atwho) {
                    n.atwho("load", "@", newVal);
                }
            },
            "atTowns": function(newVal) {
                if (!newVal) {
                    return;
                }
                r.set("atTownData", unique(r.get("atTownData").concat(newVal)));
            },
            "atTownData": function(newVal) {
                if (!newVal) {
                    return;
                }
                var n = $(r.get("editNode"));
                if (n.atwho) {
                    n.atwho("load", "/town/", newVal);
                }
            },
        });

    },
};


function autocomplete(node, r, destroy) {
    "use strict";

    if (destroy) {
        $(node).atwho('destroy');
        return;
    }

    var emojiList = $.map(ts.emojiList, function(emoji, key) {
        return key.slice(1, key.length - 1);
    }).reverse();

    var atWhoNode = $(node).atwho({
        at: "@",
        data: r.get("atUserData"),
        searchKey: "username",
        displayTpl: "<li>${username}  <small>(${name})</small> </li>",
        insertTpl: "@${username}",
    }).atwho({
        at: ":",
        data: emojiList,
        displayTpl: function(value) {
            return "<li>" + emojione.shortnameToImage(":" + value.name + ":") + value.name + "</li>";
        },
        insertTpl: ":${name}:",
    }).atwho({
        at: "#",
        data: ts.emojiCategory("tags"),
        searchKey: "tag",
        displayTpl: function(value) {
            return "<li><img src='/images/tags/${tag}.png' title='tagged - ${name}'> ${tag} <small>${name}</small></li>";
        },
        insertTpl: "#${tag}",
    }).atwho({
        at: "/town/",
        data: r.get("atTownData"),
        insertTpl: "/town/${name}",
    }).on("inserted.atwho", function() {
        //Ractive doesn't seem to capture when atWho
        // updates the textarea, so I need to manually update
        // it here off of the select event
        r.set("input", $(this).val());
    });

}
