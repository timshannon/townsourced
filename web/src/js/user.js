// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

//Ractive + components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Modal from "./components/modal";
import ImageUpload from "./components/imageUpload";
import PreviewEditor from "./components/previewEditor";
import UserIcon from "./components/userIcon";
import Toggle from "./components/toggle";
import ButtonGroup from "./components/buttonGroup";
import {
    scale
}
from "./lib/ractive-transition-scale";

//3rd party libs
import "./lib/bootstrap/tab";
import "./lib/bootstrap/dropdown";

//ts libs
import {
    err
}
from "./ts/error";
import * as User from "./ts/user";
import {
    isModerator,
    loadAndStoreTown,
}
from "./ts/town";
import {
    htmlPayload,
    pluck,
    addPager,
    urlJoin,
    since,
    formatDate,
    urlQuery,
}
from "./ts/util";

import {
    categories as postCategories,
    statuses as postStatuses,
    isModerated as postIsModerated,
    loadAndStorePosts,
}
from "./ts/post";

import Facebook from "./ts/facebook";
import Google from "./ts/google";
import Twitter from "./ts/twitter";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            modal: Modal,
            imageUpload: ImageUpload,
            previewEditor: PreviewEditor,
            userIcon: UserIcon,
            toggle: Toggle,
            buttonGroup: ButtonGroup,
        },
        transitions: {
            scale: scale
        },
        data: function() {
            return {
                user: htmlPayload(),
                currentPost: {},
                postCategories: postCategories,
                postStatuses: $.extend({
                    "all": "All"
                }, postStatuses),
                users: {},
                notifications: {
                    pageSize: 15,
                    page: 1,
                    data: [],
                    fetch: function(last) {
                        getNotifications(last.when);
                    },
                },
                townView: "All",
                towns: {
                    pageSize: 15, //towns per page
                    page: 1,
                    data: [],
                },
                postStatus: "all",
                posts: {
                    pageSize: 15, //posts per page
                    page: 1,
                    data: [],
                    fetch: function(last, len) {
                        getPosts(last.updated, len);
                    },
                },
                postIsModerated: postIsModerated,
                comments: {
                    pageSize: 15, //comments per page
                    page: 1,
                    data: [],
                    fetch: function(last) {
                        getComments(last.updated);
                    },
                },
                processComment: function(comment) {
                    if (!comment) {
                        return;
                    }
                    return ts.processMessage(comment.substring(0, 1000));
                },
                since: since,
                formatDate: formatDate,
            };
        },
        onconfig: function() {
            addPager(this);
        },
    });

    if (r.get("user.self")) {
        r.set("tab", "Notifications");
        r.set("notificationView", "Unread");
        r.set("postView", "saved");
    } else {
        r.set("postView", "my");
        r.set("tab", "Posts");
    }


    setUser(r.get("user"));

    //ractive events
    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser) {
                r.set("currentUser", currentUser);
            }
        },
        "nameModal": function(event) {
            event.original.preventDefault();
            $("#nameModal").on("shown.bs.modal", function() {
                $("#nameChange").focus();
            });

            if (r.get("noName")) {
                r.set("nameChange", "");
            } else {
                r.set("nameChange", r.get("user.name"));
            }

            r.set("nameErr", false);

            $("#nameModal").modal();
        },
        "nameSave": function(event) {
            event.original.preventDefault();
            User.setName(event.context.nameChange, r.get("user.vertag"))
                .done(function() {
                    getUser(r.get("user.username"));
                    $("#nameModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("nameErr", err(result).message);
                });
        },
        "emailModal": function(event) {
            event.original.preventDefault();
            $("#emailModal").on("shown.bs.modal", function() {
                $("#emailChange").focus();
            });

            r.set("emailChange", r.get("user.email"));

            r.set("emailErr", false);
            r.set("emailChangePassword", "");

            $("#emailModal").modal();
        },
        "emailSave": function(event) {
            event.original.preventDefault();
            if (r.get("user.email") == event.context.emailChange) {
                $("#emailModal").modal("hide");
                return;
            }

            User.validateEmail(event.context.emailChange)
                .fail(function(result) {
                    r.set("emailErr", result.message);
                })
                .done(function() {
                    User.setEmail(event.context.emailChange, r.get("user.vertag"), event.context.emailChangePassword)
                        .done(function() {
                            getUser(r.get("user.username"));
                            $("#emailModal").modal("hide");
                        })
                        .fail(function(result) {
                            r.set("emailErr", err(result).message);
                        });
                });
        },
        "passwordModal": function(event) {
            event.original.preventDefault();
            $("#passwordModal").on("shown.bs.modal", function() {
                $("#passwordCurrent").focus();
            });

            r.set("passwordCurrent", "");
            r.set("passwordNew", "");
            r.set("passwordNew2", "");

            r.set("passwordErr", false);
            r.set("password2Err", false);

            $("#passwordModal").modal();
        },
        "passwordSave": function(event) {
            event.original.preventDefault();
            if (event.context.passwordNew !== event.context.passwordNew2) {
                r.set("password2Err", "Passwords do not match");
                return;
            }
            User.setPassword(event.context.passwordCurrent, event.context.passwordNew, r.get("user.vertag"))
                .done(function() {
                    getUser(r.get("user.username"));
                    $("#passwordModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("passwordErr", err(result).message);
                });
        },
        "toggleFacebook": function(event) {
            if (r.get("user.facebookID")) {
                Facebook.disconnect()
                    .done(function() {
                        getUser(r.get("user.username"));
                    })
                    .fail(function(result) {
                        if (err(result)) {
                            window.location = "/error/";
                        }
                    });

            } else {
                Facebook.startLogin();
            }
        },
        "toggleGoogle": function(event) {
            if (r.get("user.googleID")) {

                Google.disconnect()
                    .done(function() {
                        getUser(r.get("user.username"));
                    })
                    .fail(function(result) {
                        if (err(result)) {
                            window.location = "/error/";
                        }
                    });
            } else {
                Google.startLogin();
            }
        },
        "toggleTwitter": function(event) {
            if (r.get("user.twitterID")) {
                Twitter.disconnect()
                    .done(function() {
                        getUser(r.get("user.username"));
                    })
                    .fail(function(result) {
                        if (err(result)) {
                            window.location = "/error/";
                        }
                    });
            } else {
                Twitter.startLogin();
            }
        },
        "toggleNotifyPost": function() {
            toggleSetting("notifyPost");
        },
        "toggleNotifyComment": function() {
            toggleSetting("notifyComment");
        },
        "toggleEmailPrivateMsg": function() {
            toggleSetting("emailPrivateMsg");
        },
        "toggleEmailPostMention": function() {
            toggleSetting("emailPostMention");
        },
        "toggleEmailCommentMention": function() {
            toggleSetting("emailCommentMention");
        },
        "toggleEmailCommentReply": function() {
            toggleSetting("emailCommentReply");
        },
        "toggleEmailPostComment": function() {
            toggleSetting("emailPostComment");
        },
        "notifications": function(event) {
            r.set("tab", "Notifications");
            getNotifications();
        },
        "posts": function(event) {
            r.set("tab", "Posts");
            getPosts();
        },
        "selectPostStatus": function(event, status) {
            event.original.preventDefault();
            r.set("postStatus", status);
            getPosts();
        },
        "comments": function(event) {
            r.set("tab", "Comments");
            getComments();
        },
        "towns": function(event) {
            r.set("tab", "Towns");
            r.set("tLoading", true);
            r.set("towns.page", 1);
            r.set("modTowns", []);

            User.towns(r.get("user.username"))
                .done(function(result) {
                    var towns = result.data || [];
                    for (var i = 0; i < towns.length; i++) {
                        if (isModerator(towns[i], r.get("user"))) {
                            towns[i].mod = true;
                            r.push("modTowns", towns[i]);
                        }
                    }
                    r.set("allTowns", towns);

                    if (r.get("townView") == "All") {
                        r.set("towns.data", r.get("allTowns"));
                    } else {
                        r.set("towns.data", r.get("modTowns"));
                    }
                })
                .fail(function(result) {
                    r.set("townErr", err(result).message);
                })
                .always(function() {
                    r.set("tLoading", false);
                });
        },
        "confirmEmail": function(event) {
            event.original.preventDefault();
            User.confirmEmail()
                .done(function() {
                    r.set("confirmEmailSent", true);
                });
        },
        "settings": function(event) {
            r.set("tab", "Settings");
        },
        "imageModal": function(event) {
            event.original.preventDefault();
            $("#imageModal").modal();
            r.findComponent("imageUpload").fire("reset");
        },
        "imageSet": function(event) {
            event.original.preventDefault();
            var img = r.get("newProfileImage");
            User.setProfileImage(img.key, img.x0, img.y0, img.x1, img.y1, r.get("user.vertag"))
                .done(function() {
                    location.reload();
                })
                .fail(function(result) {
                    r.findComponent("imageUpload").set("error", err(result).message);
                });
        },
        "viewNotification": function(event, keypath) {
            var context;

            if (event) {
                event.original.preventDefault();
                context = event.context;
                keypath = event.keypath;
            } else {
                context = r.get(keypath);
            }

            r.set("messageOriginal", context.message);
            r.set("message", ts.processMessage(context.message));
            r.set("messageSubject", context.subject);
            r.set("messageFrom", r.get("users." + context.from));
            if (r.get("notificationView") != "Sent") {
                r.set("messageUser", r.get("users." + context.from));
                r.fire("markAsRead", event, context, keypath);
            } else {
                r.set("messageUser", r.get("users." + context.username));
            }
            $("#notificationModal").modal();
        },
        "markAllRead": function(event) {
            event.original.preventDefault();
            User.markAllNotificationsRead()
                .done(function() {
                    r.fire("notifications");
                    r.findComponent("navbar").fire("checkNotifications");
                })
                .fail(function(result) {
                    r.set("notifications.data", []);
                    r.set("notificationErr", err(result).message);
                });
        },
        "markAsRead": function(event, context, keypath) {
            if (event) {
                event.original.preventDefault();
                context = event.context;
                keypath = event.keypath;
            }

            if (context.read) {
                return;
            }

            context.read = true;
            r.set(keypath, context);

            User.markNotificationRead(context.key)
                .done(function() {
                    r.findComponent("navbar").fire("checkNotifications");
                })
                .fail(function(result) {
                    r.fire("notifications");
                });
        },
        "sendDM": function() {
            $("#sendMessageModal").on("shown.bs.modal", function() {
                $("#dmSubject").focus();
            });

            r.set("dmTo", r.get("user"));

            r.set("atUsers", [r.get("user"), r.get("currentUser")]);
            r.set("dmSubject", null);
            r.set("dmMessage", null);
            r.set("dmErr", false);

            $("#sendMessageModal").modal();

        },
        "reply": function(event) {
            $("#notificationModal").modal("hide");
            $("#sendMessageModal").on("shown.bs.modal", function() {
                r.findComponent("editor").fire("setSelection", 0);
            });

            r.set("dmTo", event.context.messageFrom);
            r.set("atUsers", [event.context.messageFrom, User.current]);
            r.set("atTowns", pluck(User.current.townKeys, "key"));

            r.set("dmSubject", "RE: " + event.context.messageSubject);
            r.set("dmMessage", makeReply(event.context.messageOriginal, event.context.messageFrom));
            r.set("dmErr", false);

            $("#sendMessageModal").modal();
        },
        "sendMessage": function(event) {
            event.original.preventDefault();
            r.set("dmSending", true);
            User.sendMessage(r.get("dmTo.username"), r.get("dmSubject"), r.get("dmMessage"))
                .done(function() {
                    $("#sendMessageModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("dmErr", err(result).message);
                })
                .always(function() {
                    r.set("dmSending", false);
                });
        },
        "viewPostModeration": function(event) {
            event.original.preventDefault();
            //Load town info
            r.set("currentPost", event.context);
            var moderation = event.context.moderation;
            for (var i = 0; i < moderation.length; i++) {
                User.loadAndStoreUser(r, moderation[i].who, "users");
                loadAndStoreTown(r, moderation[i].town, null, r.get("towns.data"));
            }
            $("#postModModal").modal();
        },
    });


    //observations
    r.observe({
        "notificationView": function() {
            getNotifications();
        },
        "townView": function(newVal) {
            if (r.get("towns.data")) {
                if (newVal == "All") {
                    r.set("towns.data", r.get("allTowns"));
                } else {
                    r.set("towns.data", r.get("modTowns"));
                }
            }
        },
        "postView": function(newVal) {
            r.set("postStatus", "all");
            getPosts();
        },
    });

    var hash = window.location.hash.substr(1);
    if (hash == "settings") {
        $('.nav-tabs a[href="#settings"]').tab('show');
        r.fire("settings");
    } else if (hash == "posts") {
        $('.nav-tabs a[href="#posts"]').tab('show');
        r.fire("posts");
    } else if (hash == "towns") {
        $('.nav-tabs a[href="#towns"]').tab('show');
        r.fire("towns");
    } else if (hash == "comments") {
        $('.nav-tabs a[href="#comments"]').tab('show');
        r.fire("comments");
    }



    //functions

    function getUser(username) {
        User.get(username)
            .done(function(result) {
                setUser(result.data);
            })
            .fail(function(result) {
                err(result, true);
            });

    }

    function getPosts(since, from) {
        r.set("pLoading", true);
        if (!since && !from) {
            r.set("posts.data", []);
            r.set("posts.page", 1);
        }

        var limit = 50;

        var call;

        if (r.get("postView") == "saved") {
            call = User.savedPosts(r.get("user.username"), r.get("postStatus"), from, limit);
        } else {
            call = User.posts(r.get("user.username"), r.get("postStatus"), since, limit);
        }

        call.done(function(result) {
                if (result.data) {
                    r.set("posts.data", r.get("posts.data").concat(result.data));
                }
            })
            .fail(function(result) {
                r.set("posts.data", []);
                r.set("postsErr", err(result).message);
            })
            .always(function() {
                r.set("pLoading", false);
            });
    }

    function getNotifications(since) {
        r.set("nLoading", true);
        if (!since) {
            r.set("notifications.data", []);
            r.set("notifications.page", 1);
        }

        var func;
        if (r.get("notificationView") === "All") {
            func = User.allNotifications(since, 50);
        } else if (r.get("notificationView") === "Sent") {
            func = User.sentNotifications(since, 50);
        } else {
            func = User.unreadNotifications(since, 50);
        }
        func.done(function(result) {
                r.set("notificationErr", null);
                if (result.data) {
                    for (var i = 0; i < result.data.length; i++) {
                        if (r.get("notificationView") === "Sent") {
                            User.loadAndStoreUser(r, result.data[i].username, "users");
                        } else {
                            User.loadAndStoreUser(r, result.data[i].from, "users");
                        }

                        r.push("notifications.data", result.data[i]);
                    }
                }
            })
            .fail(function(result) {
                r.set("notifications.data", []);
                r.set("notificationErr", err(result).message);
            })
            .always(function() {
                r.set("nLoading", false);
            });

    }

    function setUser(user) {
        if (!user.name) {
            r.set("noName", true);
            user.name = user.username;
        } else {
            r.set("noName", false);
        }
        user.canToggle = {
            facebook: canToggle(user, "facebook"),
            google: canToggle(user, "google"),
            twitter: canToggle(user, "twitter"),
        };

        // self comes from server, so preserve it
        user.self = r.get("user.self");

        r.set("user", user);
        document.title = "User - " + user.name + " - townsourced";
    }

    function canToggle(user, idType) {
        var id = idType + "ID";
        if (!user[id]) {
            return true;
        }

        if (idType == "facebook" && !user.googleID && !user.twitterID && !user.hasPassword) {
            return false;
        }
        if (idType == "google" && !user.facebookID && !user.twitterID && !user.hasPassword) {
            return false;
        }
        if (idType == "twitter" && !user.facebookID && !user.googleID && !user.hasPassword) {
            return false;
        }

        return true;
    }


    function makeReply(msg, from) {
        var lines = msg.split("\n");

        var top = "\n\n\n----------------------------------------\n";
        var icon = "/images/emoji/png/1f464.png?v=1.2.4";
        if (from.profileIcon) {
            icon = urlJoin("/api/v1/user/", from.username, "/image/?icon");
        }

        icon = "> ![profile image](" + icon + " '" + from.username + "') " + lines.shift() + "\n";

        for (var i = 0; i < lines.length; i++) {
            lines[i] = "> " + lines[i];
        }

        return top + icon + lines.join("\n");
    }

    function getComments(since) {
        r.set("cLoading", true);
        if (!since) {
            r.set("comments.data", []);
            r.set("comments.page", 1);
        }

        var limit = 50;

        User.comments(r.get("user.username"), {
                since: since,
                limit: limit,
            })
            .done(function(result) {
                if (result.data) {
                    r.set("comments.data", r.get("comments.data").concat(result.data));
                }
            })
            .fail(function(result) {
                r.set("comments.data", []);
                r.set("commentsErr", err(result).message);
            })
            .always(function() {
                r.set("cLoading", false);
            });
    }

    function toggleSetting(setting) {
        var data = {
            vertag: r.get("user.vertag"),
        };
        data[setting] = (r.get("user." + setting) !== true);

        User.set(data)
            .done(function() {
                getUser(r.get("user.username"));
            })
            .fail(function(result) {
                r.set("settingsErr", err(result).message);
            });
    }

});
