// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

//Ractive
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Alert from "./components/alert";
import Header from "./components/header";
import CategorySelect from "./components/categorySelect";
import PostList from "./components/postList";
import Modal from "./components/modal";
import TownPanel from "./components/townPanel";
import PostOptions from "./components/postOptions";
import {
    scale
}
from "./lib/ractive-transition-scale";

//ts
import * as Town from "./ts/town";
import {
    err
}
from "./ts/error";
import {
    buildSearchParams,
}
from "./ts/search";
import {
    htmlPayload,
    scrollToFixed,
}
from "./ts/util";


$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            alert: Alert,
            header: Header,
            categorySelect: CategorySelect,
            postList: PostList,
            modal: Modal,
            townPanel: TownPanel,
            postOptions: PostOptions,
        },
	transitions: {
            scale: scale
        },
        data: function() {
            var town = htmlPayload("townPayload");
            var currentUser = htmlPayload("userPayload");
            var townLoad = {};
            townLoad[town.key] = town;
            return {
                authenticated: false,
                town: town,
                townLoad: townLoad,
                searchOptions: {
                    towns: [town.key],
                    showModerated: false,
                },
                posts: htmlPayload("postsPayload"),
                currentUser: currentUser,
                isMember: true,
                canJoin: false,
                isMod: Town.isModerator(town, currentUser),
                canPost: Town.canPost,
                requestSent: false,
            };
        },
    });

    Town.setTheme(r.get("town.color"));
    var postList = r.findComponent("postList");

    scrollToFixed(null, r.nodes.fixedTrigger, null, function(fixed) {
        r.set("fixed", fixed);
    });

    scrollToFixed(r.nodes.navCenter, r.nodes.fixedTrigger);

    r.set("canJoin", Town.canJoin(r.get("town"), r.get("currentUser")));

    //ractive events
    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser) {
                r.set("authenticated", true);
            }
            r.set("isMember", Town.isMember(r.get("town"), currentUser));
        },
        "infoPanel": function(event) {
            event.original.preventDefault();
            if (!r.get("infoParsed")) {
                r.set("infoParsed", ts.processMessage(r.get("town.information")));
            }
            $("#infoPanel").modal();
        },
        "changeCategory": function() {
            postList.fire("getPosts");
        },
        "townPanel.join join": function(event) {
            event.original.preventDefault();
            if (!r.get("authenticated")) {
                r.findComponent("navbar").fire("login",
                    "Log in to join a town");
                return;
            }

            Town.join(r.get("town.key"))
                .done(function() {
                    location.reload();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "townPanel.leave leave": function() {
            Town.leave(r.get("town.key"))
                .done(function() {
                    location.reload();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "requestInvite": function(event) {
            event.original.preventDefault();
            if (!r.get("authenticated")) {
                r.findComponent("navbar").fire("login",
                    "Log in to request an invite");
                return;
            }

            Town.requestInvite(r.get("town.key"))
                .done(function() {
                    r.set("requestSent", true);
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });

        },
        "search": function(event) {
            event.original.preventDefault();

            if (!r.get("searchOptions.search").trim()) {
                return;
            }

            window.location = "/town/" + r.get("town.key") + "/search?" + buildSearchParams({
                search: r.get("searchOptions.search"),
                category: r.get("searchOptions.category"),
            });
        },
        "post": function(event) {
            event.original.preventDefault();
            r.findComponent("postOptions").fire("show");
        },

    });

    r.observe("searchOptions.showModerated", function() {
        postList.fire("getPosts");
    }, {
        init: false
    });

    r.observe("searchOptions.showModeratedAny", function() {
        postList.fire("reset", true);
    }, {
        init: false
    });



    //functions


});
