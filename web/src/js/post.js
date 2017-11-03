// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive 
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Modal from "./components/modal";
import PostCmp from "./components/post";
import Alert from "./components/alert";
import Comments from "./components/comments";
import UserIcon from "./components/userIcon";
import PreviewEditor from "./components/previewEditor";
import UserSelect from "./components/userSelect";

//ts libs
import * as Post from "./ts/post";
import * as Comment from "./ts/comment";
import {
    get as getTown,
    isModerator,
    loadAndStoreTown,
}
from "./ts/town";
import {
    err
}
from "./ts/error";
import {
    htmlPayload,
    since,
    formatDate,
    origin,
    urlQuery,
}
from "./ts/util";

import {
    loadAndStoreUser,
    get as getUser,
    savePost,
    removeSavedPost,
    isSavedPost,
    sendMessage,
	listName,
}
from "./ts/user";

//3rd party libs
import "./lib/bootstrap/dropdown";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            modal: Modal,
            post: PostCmp,
            alert: Alert,
            comments: Comments,
            userIcon: UserIcon,
            previewEditor: PreviewEditor,
            userSelect: UserSelect,
        },
        data: function() {
            var currentUser = htmlPayload("user");
            return {
                loaded: true,
                canModerate: false,
                modTowns: [],
                userLoad: {},
                isModerated: false,
                moderatedTowns: [],
                post: htmlPayload("post"),
                comments: htmlPayload("comments"),
                more: htmlPayload("more"),
                commentContext: htmlPayload("commentContext"),
                currentUser: currentUser,
                reportReasons: Post.reportReasons,
                reason: Post.reportReasons[0],
                isCreator: false,
                categories: Post.categories,
                since: since,
                formatDate: formatDate,
                commentsLoading: true,
                saved: function(post) {
                    if (!post) {
                        return false;
                    }
                    return isSavedPost(this.get("currentUser"), post.key);
                },
                editTimeRemain: null,
                editDuration: Post.editDuration,
                facebookLink: function(appID) {
                    return "https://www.facebook.com/dialog/share?" +
                        $.param({
                            app_id: appID,
                            display: "popup",
                            href: window.location,
                        }, true);
                },
                twitterLink: function() {
                    var tags = this.get("post.hashTags") || [];
                    return "https://twitter.com/share?" +
                        $.param({
                            url: window.location,
                            hashtags: tags.join(","),
                            text: "Check out this post on townsourced: " + this.get("post.title"),
                        }, true);
                },
                googleLink: function() {
                    return "https://plus.google.com/share?" +
                        $.param({
                            url: window.location,
                        }, true);
                },
                townContext: null,
                contactErr: null,
                contactSubject: null,
                contactMessage: null,
				contactLoading: false,
				listName: listName,
            };
        },
        oncomplete: function() {
            var r = this;

            r.set("commentsLoading", false).then(function() {
                if (window.location.hash == "#comments") {
                    var comments = r.findComponent("comments");
                    if (comments) {
                        comments.fire("scrollToComments");
                    }
                }
            });
        },
    });

    loadAndStoreUser(r, r.get("post.creator"));
    loadAndStoreTown(r, r.get("post.townKeys"));

    var post = r.findComponent("post");

    post.fire("parse");

    editCountDown();

    r.set("townContext", urlQuery().town);


    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser && currentUser.username == r.get("post.creator")) {
                r.set("isCreator", true);
            }
        },
        "reportModal": function() {
            var nav = r.findComponent("navbar");
            if (!nav.get("user")) {
                nav.fire("login", "Login to report a post");
                return;
            }

            r.set("reason", Post.reportReasons[0]);
            r.set("otherReason", "");
            $("#reportModal").modal();
        },
        "report": function(event) {
            event.original.preventDefault();
            var reason = r.get("reason");
            if (reason == "other") {
                reason = r.get("otherReason");
            }
            if (!reason) {
                r.set("reportErr", "You must specify a reason for reporting this post");
                return;
            }
            Post.report(r.get("post.key"), reason, r.get("post.vertag"))
                .done(function(result) {
                    window.location.reload();
                })
                .fail(function(result) {
                    r.set("reportErr", err(result).message);
                });
        },
        "reportHide": function() {
            r.set("reportErr", null);
        },
        "close": function(event) {
            event.original.preventDefault();
            Post.close(r.get("post.key"), r.get("post.vertag"))
                .done(function() {
                    window.location.reload();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "reopen": function(event) {
            event.original.preventDefault();
            Post.reopen(r.get("post.key"), r.get("post.vertag"))
                .done(function() {
                    window.location.reload();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "allTowns": function(event) {
            event.original.preventDefault();
            r.toggle("allTowns");
        },
        "toggleSave": function(event) {
            event.original.preventDefault();
            var postKey = r.get("post.key");
            var saved = isSavedPost(r.get("currentUser"), postKey);

            var func;
            if (!saved) {
                func = savePost;
            } else {
                func = removeSavedPost;
            }

            func(postKey)
                .always(function() {
                    getUser()
                        .done(function(result) {
                            r.set("currentUser", result.data);
                        });
                });
        },
        "setNotifyOnComment": function(event) {
            event.original.preventDefault();
            r.toggle("post.notifyOnComment");
            Post.setNotifyOnComment(r.get("post.key"), r.get("post.vertag"), r.get("post.notifyOnComment"))
                .done(function() {
                    window.location.reload();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "unpublish": function(event) {
            event.original.preventDefault();
            Post.unpublish(r.get("post.key"), r.get("post.vertag"))
                .done(function() {
                    window.location = "/editpost/" + r.get("post.key");
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "disabled": function(event) {
            event.original.preventDefault();
        },
        "showSharePost": function(event) {
            var nav = r.findComponent("navbar");
            if (!nav.get("user")) {
                nav.fire("login", "Login to share a post on townsourced");
                return;
            }
            var lnkEnd = '](/post/' + r.get("post.key") + ')';
            var title = ts.escapeMessage(r.get("post.title"));

            r.set("shareSubject", 'Check out this post "' + r.get("post.title") + '"');
            var msg = 'Check out this post: [' + title + lnkEnd;
            if (r.get("post.featuredImage")) {
                msg += '\n\n---------\n' + '[![' + title + '](/api/v1/image/' + r.get("post.featuredImage") +
                    '?thumb)' + lnkEnd;
            }

            r.set("shareMessage", msg);
            r.set("shareWho", null);
            r.set("shareError", null);

            $("#sharePost").on("shown.bs.modal", function() {
                $("#shareWho").focus();
            });
            $("#sharePost").modal();
        },
        "sharePost": function(event) {
            event.original.preventDefault();
            r.set("shareLoading", true);
            sendMessage(r.get("shareWho.username"), r.get("shareSubject"), r.get("shareMessage"))
                .done(function(result) {
                    $("#sharePost").modal("hide");
                })
                .fail(function(result) {
                    r.set("shareError", err(result).message);
                })
                .always(function() {
                    r.set("shareLoading", false);
                });
        },
        "shareWindow": function(event, href, height, width) {
            height = height || 600;
            width = width || 600;
            window.open(href, '', 'menubar=no,toolbar=no,resizable=yes,scrollbars=yes,height=' + height +
                ',width=' + width);
            return false;
        },
        "moderateModal": function() {
            r.set("modReason", Post.reportReasons[0]);
            r.set("otherModReason", "");
            r.set("moderateErr", null);
            $("#moderateModal").modal();
        },
        "moderate": function(event) {
            event.original.preventDefault();
            var reason = r.get("modReason");
            if (reason == "other") {
                reason = r.get("otherModReason");
            }
            if (!reason) {
                r.set("moderateErr", "You must specify a reason for moderating this post");
                return;
            }
            Post.moderate(r.get("post.key"), r.get("townContext"), reason, r.get("post.vertag"))
                .done(function(result) {
                    window.location.reload();
                })
                .fail(function(result) {
                    r.set("moderateErr", err(result).message);
                });
        },
        "removeModeration": function(event) {
            Post.removeModeration(r.get("post.key"), r.get("townContext"), r.get("post.vertag"))
                .always(function(result) {
                    window.location.reload();
                });
        },
        "contactPosterModal": function(event) {
            event.original.preventDefault();
            var nav = r.findComponent("navbar");
            if (!nav.get("user")) {
                nav.fire("login", "Login to contact the poster");
                return;
            }
            var lnkEnd = '](/post/' + r.get("post.key") + ')';
            var title = ts.escapeMessage(r.get("post.title"));

            r.set("contactSubject", 'In regards to the post "' + r.get("post.title") + '"');
            var msg = 'In regards to the post: [' + title + lnkEnd;
            if (r.get("post.featuredImage")) {
                msg += '\n\n---------\n' + '[![' + title + '](/api/v1/image/' + r.get("post.featuredImage") +
                    '?thumb)' + lnkEnd;
            }

            r.set("contactMessage", msg);
            r.set("contactError", null);

            $("#contactModal").on("shown.bs.modal", function() {
                $("#contactSubject").focus();
            });
            $("#contactModal").modal();
        },
        "contactPoster": function(event) {
            event.original.preventDefault();
            r.set("contactLoading", true);
            sendMessage(r.get("post.creator"), r.get("contactSubject"), r.get("contactMessage"))
                .done(function(result) {
                    $("#contactModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("contactError", err(result).message);
                })
                .always(function() {
                    r.set("contactLoading", false);
                });
        }
    });

    if (r.get("townContext")) {
        r.observe("townLoad." + r.get("townContext"), function(newval) {
            if (!newval || newval.loading) {
                return;
            }
            if (!r.get("isModerated") && Post.isModerated(r.get("post"), r.get("townContext"))) {
                r.set("isModerated", true);
            }
            if (!r.get("canModerate")) {
                if (isModerator(newval, r.get("currentUser"))) {
                    r.set("canModerate", true);
                    return;
                } else {
                    //check if moderated
                    if (r.get("isModerated")) {
                        //post is moderated from this town context, and user isn't a moderator
                        r.findComponent("page").fire("four04");
                    }
                }
            }
        });
    }


    function editCountDown() {
        var countdown = function() {
            var published = new Date(r.get("post.published"));
            if (!published) {
                return;
            }
            var left = Math.floor((Post.editDuration - (Date.now() - published)) / 1000);
            if (left <= 0) {
                r.set("editCountDown", null);
            } else {
                r.set("editCountDown", left);
                window.setTimeout(countdown, 1000);
            }
        };
        window.setTimeout(countdown, 1000);
    }

});
