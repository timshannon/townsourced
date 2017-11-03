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
import Modal from "./components/modal";
import Map from "./components/map";
import PostOptions from "./components/postOptions";

// ts libs
import {
    htmlPayload,
    scrollToFixed,
}
from "./ts/util";
import {
    getSearchOptionsFromURL,
}
from "./ts/search";

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
            header: Header,
            categorySelect: CategorySelect,
            searchSidebar: SearchSidebar,
            modal: Modal,
            map: Map,
            postOptions: PostOptions,
        },
        data: function() {
            return {
                category: "all",
                posts: htmlPayload("postsPayload"),
                currentUser: htmlPayload("userPayload"),
                error: err(htmlPayload("errorPayload")).message,
                srcOptions: {
                    category: "all",
                    search: "",
                    tags: [],
                    milesDistant: 3,
                    sort: "none",
                },
            };
        },
    });


    r.set("srcOptions", getSearchOptionsFromURL());
    scrollToFixed(r.nodes.navCenter, r.nodes.fixedTrigger);

    r.on({
        "locationModal": function() {
            $("#locationModal").on("shown.bs.modal", function() {
                var map = r.findComponent("map");
                if (map) {
                    map.fire("reset");
                    if (r.get("srcOptions.latitude") && r.get("srcOptions.longitude")) {
                        map.fire("setLocation", r.get("srcOptions.latitude"), r.get("srcOptions.longitude"), 12);
                    }
                }
            });

            $("#locationModal").modal();
        },
        "setLocation": function() {
            r.set("srcOptions.latitude", r.get("currentLatitude"));
            r.set("srcOptions.longitude", r.get("currentLongitude"));
            r.set("srcOptions.milesDistant", r.get("currentMilesDistant"));
            r.set("srcOptions.tags", []);
            r.set("showList", false);
            $("#locationModal").modal("hide");
            if (r.get("srcOptions.search")) {
                r.set("showList", true);
                search();
            } else {
                $("#search").focus();
            }
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
        "search": function() {
            search();
        },
        "searchInput": function(event) {
            event.original.preventDefault();

            if (!r.get("srcOptions.search").trim()) {
                return;
            }
            r.set("showList", true);
            search();
        },
        "myLocation": function() {
            if ("geolocation" in navigator) {
                navigator.geolocation.getCurrentPosition(function(position) {
                    r.set("srcOptions.latitude", position.coords.latitude);
                    r.set("srcOptions.longitude", position.coords.longitude);
                    if (r.get("srcOptions.search")) {
                        r.set("showList", true);
                        search();
                    } else {
                        $("#search").focus();
                    }

                }, function(error) {
                    r.set("error", "Sorry, we were unable to get your location: " + error.message);
                });
            } else {
                r.set("error", "We cannot get your current location from this browser, sorry.");
            }
        },
		"post": function(event) {
			event.original.preventDefault();
            r.findComponent("postOptions").fire("show");
        },
    });

    function search() {
        var postList = r.findComponent("postList");
        if (postList) {
            postList.fire("getPosts");
        }
    }

});
