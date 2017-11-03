// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive + components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import PostList from "./components/postList";
import TownList from "./components/townList";
import Header from "./components/header";
import Alert from "./components/alert";
import CategorySelect from "./components/categorySelect";
import TownPanel from "./components/townPanel";
import Modal from "./components/modal";
import PostOptions from "./components/postOptions";


// ts libs
import {
    setTheme,
    join,
    leave,
}
from "./ts/town";
import {
    htmlPayload,
    pluck,
    scrollToFixed,
}
from "./ts/util";
import {
    buildSearchParams,
}
from "./ts/search";

import {
    towns as getUserTowns,
}
from "./ts/user";

import {
    err
}
from "./ts/error";


//3rd party


$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            postList: PostList,
            alert: Alert,
            townList: TownList,
            header: Header,
            categorySelect: CategorySelect,
            modal: Modal,
            townPanel: TownPanel,
            postOptions: PostOptions,
        },
        data: function() {
            var towns = htmlPayload("townsPayload") || [];
            return {
                searchOptions: {
                    category: "all",
                    towns: [],
                },
                posts: htmlPayload("postsPayload"),
                currentUser: htmlPayload("userPayload"),
                towns: towns.slice(0, 3),
                town: null,
                error: null,
                userTowns: [],
            };
        },
    });

    var postList = r.findComponent("postList");

    scrollToFixed(r.nodes.navCenter, r.nodes.fixedTrigger);
    scrollToFixed(null, r.nodes.fixedTrigger, null, function(fixed) {
        r.set("fixed", fixed);
    });


    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser) {
                userTowns();
                return;
            }
        },
        "search": function(event) {
            event.original.preventDefault();

            if (!r.get("searchOptions.search").trim()) {
                return;
            }
            window.location = "/search?" + buildSearchParams({
                search: r.get("searchOptions.search"),
                category: r.get("searchOptions.category"),
            });
        },
        "changeCategory townList.select townList.deselect": function(event) {
            postList.fire("getPosts");
        },
        "selectTown": function(event) {
            event.original.preventDefault();
            r.set("town", event.context);

            setTheme(r.get("town.color"));
            $("#infoPanel").modal();
        },
        "townPanel.join": function(event) {
            event.original.preventDefault();
            join(r.get("town.key"))
                .done(function(result) {
                    location.reload();
                })
                .fail(function(result) {
                    r.set("townError", err(result).message);
                });
        },
        "townPanel.leave": function(event) {
            event.original.preventDefault();
            leave(r.get("town.key"))
                .done(function(result) {
                    location.reload();
                })
                .fail(function(result) {
                    r.set("townError", err(result).message);
                });
        },
        "post": function(event) {
            event.original.preventDefault();
            r.findComponent("postOptions").fire("show");
        },
    });

    function userTowns() {
        getUserTowns()
            .done(function(result) {
                r.set("userTowns", result.data);
            })
            .fail(function(result) {
                r.set("error", err(result).message);
            });
    }


});
