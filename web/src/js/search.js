// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive + components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import PostList from "./components/postList";
import Header from "./components/header";
import Alert from "./components/alert";
import CategorySelect from "./components/categorySelect";
import SearchSidebar from "./components/searchSidebar";
import PostOptions from "./components/postOptions";

// ts libs
import {
    loadAndStoreTown,
    setTheme,
    isModerator,
}
from "./ts/town";
import {
    htmlPayload,
    pluck,
    unique,
    scrollToFixed,
}
from "./ts/util";
import {
    err
}
from "./ts/error";

import {
    buildSearchParams,
    getSearchOptionsFromURL,
}
from "./ts/search";

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
            header: Header,
            categorySelect: CategorySelect,
            searchSidebar: SearchSidebar,
            postOptions: PostOptions,
        },
        data: function() {
            var town = htmlPayload("townPayload");
            var currentUser = htmlPayload("userPayload");

            return {
                posts: htmlPayload("postsPayload"),
                currentUser: currentUser,
                town: town,
                error: err(htmlPayload("errorPayload")).message,
                srcOptions: {
                    category: "all",
                    search: "",
                    tags: [],
                    towns: [],
                    sort: "none", showModerated: false,
                },
                buildSearchParams: buildSearchParams,
                isMod: isModerator(town, currentUser),
            };
        },
    });

    r.set("srcOptions", getSearchOptionsFromURL());

    scrollToFixed(r.nodes.navCenter, r.nodes.fixedTrigger);
    scrollToFixed(null, r.nodes.fixedTrigger, null, function(fixed) {
        r.set("fixed", fixed);
    });


    if (r.get("town")) {
        r.set("townLoad." + r.get("town.key"), r.get("town"));
        r.set("srcOptions.towns", unique([r.get("town.key")].concat(r.get("srcOptions.towns") || [])));
        setTheme(r.get("town.color"));
    }


    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser) {
                loadAndStoreTown(r, pluck(currentUser.townKeys, "key"));
                return;
            }
        },
        "search": function() {
            search();
        },
        "changeCategory": function() {
            var postList = r.findComponent("postList");
            if (postList) {
                postList.fire("getPosts");
            }
        },
        "searchSidebar.hidden searchSidebar.shown": function() {
            r.findComponent("postList").fire("reset", true);
        },
        "searchInput": function(event) {
            event.original.preventDefault();

            if (!r.get("srcOptions.search").trim()) {
                return;
            }
            search();
        },
"post": function(event) {
			event.original.preventDefault();
            r.findComponent("postOptions").fire("show");
        },
    });

    function search() {
        var srcOptions = r.get("srcOptions");

        if (r.get("town")) {
            srcOptions.towns = [];
            window.location = "/town/" + r.get("town.key") + "/search?" + buildSearchParams(srcOptions);
        } else {
            window.location = "/search?" + buildSearchParams(srcOptions);
        }
    }

});
