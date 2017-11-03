// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

//Ractive & components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Editor from "./components/editor";
import ImageUpload from "./components/imageUpload";
import ImageZoom from "./components/imageZoom";
import Gallery from "./components/gallery";
import PostCmp from "./components/post";
import ButtonGroup from "./components/buttonGroup";
import TownList from "./components/townList";
import Alert from "./components/alert";

// TS libs
import * as Post from "./ts/post";
import {
    err
}
from "./ts/error";
import {
    canPost,
    announcementTown,
	loadAndStoreTown,
}
from "./ts/town";

import {
    htmlPayload,
    urlQuery,
    unique,
    pluck,
}
from "./ts/util";

import {
    towns as userGetTowns,
}
from "./ts/user";

import {
    bookmarklet,
}
from "./ts/bookmarklet";

import {
    pushUniq as storagePushUniq,
}
from "./ts/storage";


// 3rd party libs
import "./lib/bootstrap/tab";
import "./lib/bootstrap/dropdown";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            editor: Editor,
            imageUpload: ImageUpload,
            imageZoom: ImageZoom,
            gallery: Gallery,
            post: PostCmp,
            buttonGroup: ButtonGroup,
            townList: TownList,
            alert: Alert,
        },
        data: function() {
            var post = htmlPayload() || {
                title: "",
                format: "standard",
                images: [],
                category: "",
                allowComments: true,
                notifyOnComment: true,
            };

            if (!post.images) {
                post.images = [];
            }
            return {
                post: post,
                user: htmlPayload("userPayload"),
                error: htmlPayload("errorPayload"),
                categories: Post.categories,
                loaded: true,
                maxImages: Post.maxImages,
                towns: {},
                atTowns: [],
                maxTitle: Post.maxTitle,
                bookmarklet: bookmarklet(),
            };
        },
    });

    var query = urlQuery();

    $("#postTitle").focus();

    if (r.get("post.key")) {
        // post is loaded from htmlpayload
        document.title = "Edit Post - townsourced";
        if (!r.get("post.images")) {
            r.set("post.images", []);
        }
    } else {
        // new post
        document.title = "Submit a new post - townsourced";
    }


    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (!currentUser) {
                r.findComponent("page").fire("notAuthenticated",
                    "You must be logged in to submit or edit a post");
                return;
            }

            if (query.town) {
                r.push("post.townKeys", query.town);
            }

            r.set("post.townKeys", unique(r.get("post.townKeys")));

            addTowns(r.get("post.townkeys"));

            if (r.get("towns").length === 0) {
                r.findComponent("page").fire("four04");
                return;
            }

            r.set("atUsers", [currentUser]);
        },
        "cancelEvent": function(event) {
            event.original.preventDefault();
        },
        "preview": function() {
            r.findComponent("post").fire("parse");
        },
        "imageUpload.uploadComplete": function(image) {
            r.findComponent("imageUpload").fire("reset");
            r.unshift("post.images", image.key);
        },
        "setCategory": function(event, category) {
            event.original.preventDefault();
            if (r.get("catMissing")) {
                r.set("catMissing", false);
                r.set("error", false);
            }
            r.set("post.category", category);
        },
        "townList.select": function(event) {
            if (r.get("townMissing")) {
                r.set("townMissing", false);
                r.set("error", false);
            }
        },
        "dismissAlert": function() {
            r.set("error", null);
        },
        "saveDraft": function() {
            if (!validate(true)) {
                return;
            }
            var post = r.get("post");
            r.set("saving", true);

            storagePushUniq("towns", r.get("post.townKeys"));

            var func;
            if (post.key) {
                func = Post.save(post);
            } else {
                func = Post.newPost(post, true);
            }

            func.done(function(result) {
                    setSaved(true);

                    if (result.data && result.data.key) {
                        window.location = "/editpost/" + result.data.key;
                    } else {
                        window.location.reload();
                    }
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                    r.set("saving", false);
                });
        },
        "publish": function() {
            if (!validate()) {
                return;
            }
            var post = r.get("post");
            r.set("saving", true);

            storagePushUniq("towns", r.get("post.townKeys"));

            var func;
            if (post.key) {
                func = Post.publish(post);
            } else {
                func = Post.newPost(post);
            }

            func.done(function(result) {
                    var key;

                    setSaved(true);

                    if (result.data && result.data.key) {
                        key = result.data.key;
                    } else {
                        key = post.key;
                    }

                    window.location = "/post/" + key;
                })
                .fail(function(result) {
                    r.set("saving", false);
                    r.set("error", err(result).message);
                });
        },
    });

    r.observe({
        "towns.*": function(newval) {
            if (newval) {
                if (!canPost(newval)) {
                    delete r.get("towns")[newval.key];
                }
            }
        },
        "post.allowComments": function(newval) {
            r.set("post.notifyOnComment", newval);
        },
    });

    r.observe("post.*", function(newval) {
        setSaved(false);
    }, {
        init: false
    });


    function setSaved(saved) {
        if (saved) {
            window.onbeforeunload = null;
            return;
        } else {
            if (window.onbeforeunload) {
                return;
            }

            window.onbeforeunload = function(e) {
                // If we haven't been passed the event get the window.event
                e = e || window.event;

                var message = "Your Post is not yet saved.";

                if (e) {
                    e.returnValue = message;
                }

                return message;
            };
        }
    }


    function addTowns(otherTownKeys) {
        userGetTowns()
            .done(function(result) {
                for (var i = 0; i < result.data.length; i++) {
                    r.set("towns." + result.data[i].key, result.data[i]);
                }

                loadAndStoreTown(r, otherTownKeys, "towns");
            })
            .fail(function(result) {
                r.set("error", err(result).message);
            });
    }

    function validate(draft) {
        if (!r.get("post.title").trim()) {
            r.set("error", "A Title is required on a post");
            return false;
        }
        if (!r.get("post.content").trim()) {
            r.set("error", "A post cannot be empty, and must contain some text in the body");
            return false;
        }
        if (!r.get("post.category") && !draft) {
            r.set("error", "Please choose a category");
            r.set("catMissing", true);
            return false;
        }
        if (!draft && (!r.get("post.townKeys") || r.get("post.townKeys.length") === 0)) {
            r.set("error", "Please choose one or more towns");
            r.set("townMissing", true);
            return false;
        }
        return true;
    }


});
