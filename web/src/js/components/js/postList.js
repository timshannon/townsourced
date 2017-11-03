// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true*/

import Ractive from "ractive";
import fade from "../../lib/ractive-transitions-fade";
import {
    posts,
    search,
    resultLimit,
}
from "../../ts/search";
import {
    pluck,
    since,
}
from "../../ts/util";
import {
    isModerated,
	price,
}
from "../../ts/post";
import {
    get as getUser,
    savePost,
    removeSavedPost,
    isSavedPost,
}
from "../../ts/user";

import {
    err
}
from "../../ts/error";

import ColumnList from "../../ts/columnList";

export
default {
    isolated: true,
    data: function() {
        return {
            postWidth: 250,
            gutter: 10,
            outsideGutter: 20,
            posts: [],
            transition: false,
            search: false,
            EOF: false,
            towns: [], // cached town data
            searchOptions: {
                towns: [], //selected towns
                category: "all",
                limit: resultLimit,
            },
            loading: false,
            colOffset: [],
            sidebar: false,
            saved: function(post) {
                if (!post) {
                    return false;
                }
                return isSavedPost(this.get("user"), post.key);
            },
            since: function(date) {
                return since(date);
            },
            price: price,
            isModerated: function(post) {
                if (!this.get("town") || !post) {
                    return false;
                }
                return isModerated(post, this.get("town.key"));
            },
            isModeratedAny: function(post) {
                if (!this.get("searchOptions.showModeratedAny")) {
                    return false;
                }
                if (!this.get("town") || !post || !post.moderation) {
                    return false;
                }

                return post.moderation.length > 0;
            },
        };
    },
    transitions: {
        load: function(t) {
            if (this.get("transition")) {
                fade(t);
            } else {
                t.complete();
            }
        },
    },
    decorators: {
        columnItem: function(node) {
            var r = this;
            var colList = r.get("colList");
            if (!colList) {
                colList = new ColumnList(node.parentNode, r.get("postWidth"), r.get("gutter"),
                    r.get("outsideGutter"), r.get("sidebar"));

                r.set("colList", colList);
                setCategoryOffset(r);
            }

            var keypath = Ractive.getNodeInfo(node).keypath;
            var item = r.get(keypath);

            item.pos = colList.nextPosition(node.getBoundingClientRect().height);
            r.set(keypath, item);

            return {
                teardown: function() {}
            };
        },
    },
    oninit: function() {
        var r = this;

        var delay;

        $(window).resize(function(e) {
            if (delay) {
                window.clearTimeout(delay);
            }
            delay = window.setTimeout(function() {
                var colList = r.get("colList");

                if (colList && !colList.fitsInWindow()) {
                    r.fire("reset", true);
                }
                setCategoryOffset(r);
            }, 50);
        });
    },
    onrender: function() {
        var r = this;

        r.on({
            "reset": function(redrawOnly) {
                var colList = r.get("colList");
                if (colList) {
                    colList.sidebar = r.get("sidebar");
                    colList.reset(redrawOnly);
                }
                r.set("redraw", true).then(function() {
                    setCategoryOffset(r);
                    r.set("redraw", false);
                });
            },

            "getPosts": function() {
                fetchPosts();
            },
            "getNextPosts": function() {
                fetchPosts(true);
            },
            "categorySelect.select": function() {
                fetchPosts();
            },
            "toggleSave": function(event) {
                event.original.preventDefault();
                var saved = isSavedPost(r.get("user"), event.context.key);

                var func;
                if (!saved) {
                    func = savePost;
                } else {
                    func = removeSavedPost;
                }

                func(event.context.key)
                    .always(function() {
                        getUser()
                            .done(function(result) {
                                r.set("user", result.data);
                            });
                    });
            },
        });

        //load on scroll
        onScroll(function() {
            if (r.get("posts.length") > 0 && !r.get("loading")) {
                r.fire("getNextPosts");
            }
        });

        function fetchPosts(next) {
            r.set("postsErr", null);
            var srcOptions = $.extend({}, r.get("searchOptions"));

            if (!next) {
                r.set("transition", false);
                r.set("posts", []);
                r.fire("reset");
            } else {
                //if already at end, don't keep checking for more posts
                // unless EOF is reset
                if (r.get("EOF")) {
                    return;
                }
                r.set("transition", true);
                srcOptions.since = r.get("posts")[r.get("posts.length") - 1].published;
                srcOptions.from = r.get("posts.length");
            }
            r.set("loading", true);

            var call;

            if (r.get("search")) {
                call = search(srcOptions);
            } else {
                call = posts(srcOptions);
            }
            call.done(function(result) {
                    if (result.data) {
                        r.set("posts", r.get("posts").concat(result.data))
                            .then(function() {
                                setCategoryOffset(r);
                                if (scrollLimit() && result.data.length == r.get("searchOptions.limit")) {
                                    fetchPosts(true);
                                    return;
                                }
                                r.set("loading", false);
                                r.set("transition", false);
                            });

                        r.set("EOF", result.data.length < r.get("searchOptions.limit"));
                    } else {
                        r.set("EOF", true);
                        r.set("loading", false);
                        r.set("transition", false);
                    }

                })
                .fail(function(result) {
                    r.set("postsErr", err(result).message);
                    r.set("loading", false);
                    r.set("transition", false);
                });
        }

    }, //onrender
    oncomplete: function() {
        var r = this;
        // if there aren't enough posts to fill the page on the initial load, load more	
        if (scrollLimit() && r.get("posts.length") > 0 && !r.get("loading")) {
            this.fire("getNextPosts");
        }
    },
};

// This is a pretty ugly function, and I apologize.  I mostly blame meg.
// It's purpose is to offset the columns heights to wrap around the category selection nicely
// I've debated putting in a generic avoid these nodes into the columnList position rules, but
// we'll stick with this for now.  If it becomes impossible to maintain, and has too many side effects,
// then I'll look at doing this generically in columnList.js
function setCategoryOffset(r) {
    var colList = r.get("colList");
    if (!colList) {
        return;
    }

    var catComp = r.findComponent("categorySelect");
    var inner = catComp.nodes.categories;
    var outer = catComp.nodes.categorySelect;
    var columns = Math.min(colList.columns.length, r.get("posts.length"));

    if (r.get("sidebar") && columns < colList.columns.length && columns > 1) {
        columns++; // add extra column for sidebar	
    }
    var itemWidth = r.get("postWidth");
    var gutter = r.get("gutter");
    var colOffset = [];

    var width = inner.getBoundingClientRect().width;
    var height = outer.parentNode.getBoundingClientRect().height;

    //determine how many columns will be offset by the category list
    // minimum number of 2 columns offset unless there is only 1 column -- meg's rules
    var offsetNum = Math.max(Math.ceil(width / (itemWidth + gutter)), Math.min(columns, 2));

    for (var i = 0; i < offsetNum; i++) {
        colOffset.push(height);
    }

    var left = columns - offsetNum;

    // make sure the non-offset columns are distributed evenly on either side of the middle columns
    if (left % 2 !== 0) {
        left -= 1;
        colOffset.push(height);
        offsetNum++;
    }
    left = left / 2;
    for (i = 0; i < left; i++) {
        colOffset.unshift(0);
        colOffset.push(0);
    }

    r.set("colOffset", colOffset);


    if (columns > 1) {
        //  grow category select to fill out to nearest post border
        var newWidth = Math.min(((offsetNum * (itemWidth + gutter)) - gutter),
            window.innerWidth - (colList.outsideGutter * 2));
        outer.style.width = newWidth + "px";
    } else {
        outer.style.width = "";
    }

    if (height !== (outer.parentNode.getBoundingClientRect().height)) {
        // width change caused a height change, set offset again
        setCategoryOffset(r);
    }

}

var delay;

function onScroll(func, optDelay) {
    "use strict";
    optDelay = optDelay || 100;
    $(window).scroll(function() {
        if (scrollLimit()) {
            if (delay) {
                window.clearTimeout(delay);
            }
            delay = window.setTimeout(func, optDelay);
        }
    });
}

function scrollLimit() {
    "use strict";
    return $(window).scrollTop() >= ($(document).height() - ($(window).height() + 800));
}
