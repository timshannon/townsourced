// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
import {
    loadAndStoreUser,
}
from "../../ts/user";
import * as Comment
from "../../ts/comment";

import {
    since,
    formatDate,
    urlQuery,
}
from "../../ts/util";
import {
    err,
}
from "../../ts/error";

import {
    scale
}
from "../../lib/ractive-transition-scale";

export
default {
    isolated: true,
    transitions: {
        newComment: function(t) {
            "use strict";
            if (this.get("newCommentScale") == t.node.id) {
                scale(t);
            } else {
                t.complete();
            }
        },
    },
    data: function() {
        "use strict";
        return {
            processComment: ts.processMessage,
            since: since,
            formatDate: formatDate,
            rootReply: "",
            reply: "",
            post: {},
            userLoad: {},
            atUsers: [],
            replyTo: "",
            comments: [],
            commentContext: null,
            saving: false,
            error: null,
            replyError: null,
            currentUser: null,
            alternate: function(keypath, offset) {
                var add = offset || 0;
                return ((commentDepth(keypath) + add) % 2 === 0);
            },
            newComment: null,
            newCommentScale: null,
            sort: urlQuery().sort || "old",
        };
    },
    onrender: function() {
        "use strict";
        var r = this;

        var page = r.findContainer("page");
        var navbar;
        if (page) {
            navbar = page.findComponent("navbar");
            if (navbar) {
                r.set("currentUser", navbar.get("user"));
            }
        }

        r.on({
            "toggleCollapsed": function(event) {
                r.toggle(event.keypath + ".collapsed");
            },
            "login": function() {
                navbar.fire("login", "Login to comment");
            },
            "showReply": function(event) {
                if (!r.get("currentUser")) {
                    navbar.fire("login", "Login to comment");
                    return;
                }
                r.set("replyTo", event.context.key).then(function() {
                    $("#replyEditor").focus();
                });
            },
            "cancelReply": function() {
                r.set("replyTo", "");
                r.set("reply", "");
            },
            "rootReply": function() {
                if (r.get("rootReply").trim() === "") {
                    return;
                }
                addComment(r.get("rootReply"));
            },
            "reply": function(event) {
                if (r.get("reply").trim() === "") {
                    return;
                }

                addComment(r.get("reply"), event.keypath);
            },
            "loadMore": function(event) {
                r.set("commentContext", event.context);
                r.fire("scrollToComments");
                getComments();

            },
            "resetRoot": function() {
                r.set("commentContext", null);
                r.fire("scrollToComments");
                getComments();
            },
            "loadNext": function(event, from) {
                getComments(from);
            },
            "scrollToComments": function(node, optComplete) {
                var scrollTop;

                if (node) {
                    // offset header, and give some breathing room (abracadabra)
                    scrollTop = node.offsetTop - 100;
                } else {
                    scrollTop = r.nodes.commentsScroll.offsetTop;
                }

                $("html, body").animate({
                    scrollTop: scrollTop,
                }, {
                    duration: 300,
                    complete: optComplete,
                });
            },
            "setSort": function(event) {
                if (r.get("sort") == "old") {
                    r.set("sort", "new");
                } else {
                    r.set("sort", "old");
                }
                getComments();
            },
        });

        r.observe({
            "userLoad": function(newval) {
                if (!newval) {
                    return;
                }
                var atUsers = r.get("atUsers");
                $.each(newval, function(k, v) {
                    if (v.username && v.name) {
                        atUsers.push(v);
                    }
                });
                r.set("atUsers", atUsers);
            },
            "comments": function(newval) {
                if (!newval) {
                    return;
                }

                for (var i = 0; i < newval.length; i++) {
                    loadUsers(newval[i]);
                }
            },
            "commentContext": function(newval) {
                if (!newval) {
                    return;
                }
                loadUsers(newval);
            },
        });

        function loadUsers(comment) {
            loadAndStoreUser(r, comment.username);
            if (comment.children) {
                for (var i = 0; i < comment.children.length; i++) {
                    loadUsers(comment.children[i]);
                }
            }
        }

        // addcomment adds a new comment, scrolls to the position where the new comment will be, then
        // scales in the new comment so the user can see in context where their reply is
        function addComment(comment, keypathContext) {
            r.set("saving", true);
            var context;
            var func;
            if (keypathContext) {
                context = r.get(keypathContext);
                func = Comment.reply(r.get("post.key"), context.key, comment);
            } else {
                func = Comment.newComment(r.get("post.key"), comment);
            }
            func.done(function(result) {
                    r.set("newComment", result.data.key);
                    r.set("newCommentScale", result.data.key);
                    if (context && commentDepth(keypathContext) > Comment.maxDepth) {
                        //if comment reply is too deep to see in the current context, then move
                        // the context to the new comment's parent
                        r.set("commentContext", context);
                    }
                    if (r.get("sort") == "old") {
                        // if comments are sorted oldest first, and the new comment will be on the next page
                        // of comments, then flip the sort order to guarentee that the new comment will be on the
                        // current page
                        if (context) {
                            if (context.children && context.children.length >= Comment.resultLimit) {
                                r.set("sort", "new");
                            }
                        } else {
                            if (r.get("comments.length") >= Comment.resultLimit) {
                                r.set("sort", "new");
                            }
                        }
                    }
                    getComments()
                        .always(function() {
                            r.set("saving", false);
                            r.set("rootReply", "");
                            r.set("reply", "");
                            r.set("replyTo", "");
                            var editors = r.findAllComponents("previewEditor") || [];
                            for (var i = 0; i < editors.length; i++) {
                                editors[i].fire("reset");
                            }

                            r.fire("scrollToComments", r.nodes[r.get("newComment")], function() {
                                r.set("newComment", null).then(function() {
                                    r.set("newCommentScale", null);
                                });

                            });
                        });
                })
                .fail(function(result) {
                    if (context) {
                        r.set("replyError", err(result).message);
                    } else {
                        r.set("error", err(result).message);
                    }
                });
        }

        function getComments(from) {
            var func;
            r.set("loading", true);
            if (r.get("commentContext")) {
                func = Comment.getChildren(r.get("post.key"), r.get("commentContext.key"), {
                        from: from,
                        sort: r.get("sort")
                    })
                    .done(function(result) {
                        if (!from) {
                            r.set("commentContext.children", result.data);
                            r.set("commentContext.hasChildren", undefined);
                        } else {
                            r.set("commentContext.children", r.get("commentContext.children").concat(result.data));
                        }
                        r.set("commentContext.more", result.more);
                    })
                    .fail(function(result) {
                        r.set("error", err(result).message);
                    });

            } else {
                func = Comment.get(r.get("post.key"), {
                        from: from,
                        sort: r.get("sort")
                    })
                    .done(function(result) {
                        if (!from) {
                            r.set("comments", result.data);
                        } else {
                            r.set("comments", r.get("comments").concat(result.data));
                        }
                        r.set("more", result.more);
                    })
                    .fail(function(result) {
                        r.set("error", err(result).message);
                    });
            }

            func.always(function() {
                r.set("loading", false);
            });

            return func;
        }
    },
};

function commentDepth(keypath) {
    "use strict";
    return keypath.split("children").length;
}
