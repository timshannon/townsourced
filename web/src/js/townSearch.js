// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

//Ractive
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Alert from "./components/alert";
import Header from "./components/header";
import Modal from "./components/modal";
import TownMap from "./components/townMap";
import TownPanel from "./components/townPanel";

//ts
import {
    search as townSearch,
    setTheme,
    join,
    leave,
}
from "./ts/town";
import {
    get as getUser
}
from "./ts/user";

import {
    err
}
from "./ts/error";
import {
    htmlPayload,
    scrollToFixed,
    urlQuery,
    addPager,
    unique,
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
            townMap: TownMap,
            modal: Modal,
            townPanel: TownPanel,
        },
        data: function() {
            return {
                authenticated: false,
                srcOptions: {
                    northBounds: null,
                    southBounds: null,
                    westBounds: null,
                    eastBounds: null,
                    search: "",
                },
                towns: {
                    pageSize: 10,
                    page: 1,
                    data: [],
                    fetch: function(last, length) {
                        nextTowns(length - (r.get("mapTowns.length") || 0));
                    },
                },
                mapTowns: [],
                searchTowns: [],
                currentUser: htmlPayload("userPayload"),
                location: htmlPayload("locationPayload"),
                showResult: false,
            };
        },
        onconfig: function() {
            addPager(this);
        },
    });

    scrollToFixed(r.nodes.navCenter, r.nodes.fixedTrigger);

    var map = r.findComponent("townMap");
    var qry = urlQuery();
    if (qry.search) {
        r.set("srcOptions.search", qry.search);
        search();
    } else {
        var location = r.get("location");
        if (location.latitude || location.longitude) {
            map.fire("pickTownAtLocation", location.latitude, location.longitude);
            r.set("showResult", true);
        }
    }

    $(r.nodes.search).focus();

    //ractive events
    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser) {
                r.set("authenticated", true);
            }
        },
        "search": function() {
            search();
        },
        "navSearchInput": function(event) {
            event.original.preventDefault();
            if (!r.get("srcOptions.search").trim()) {
                return;
            }
            search();
            scrollToResult();
        },
        "searchInput": function(event) {
            event.original.preventDefault();

            if (!r.get("srcOptions.search").trim()) {
                return;
            }
            search();
        },
        "selectTown": function(event) {
            event.original.preventDefault();
            selectTown(event.context);
        },
        "townMap.townSet": function(town) {
            setTheme(town.color);
        },
        "myLocation": function() {
            r.set("searchTowns", []);
            r.set("showResult", true);
            r.set("srcOptions.search", "");
            map.fire("myLocation");
        },
        "townPanel.join": function(error) {
            join(r.get("town.key"))
                .done(function(result) {
                    refreshUser();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "townPanel.leave": function(error) {
            leave(r.get("town.key"))
                .done(function(result) {
                    refreshUser();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "townPanel.moreInfo": function() {
            $("#infoPanel").modal();
        },
        "visitTown": function(event) {
            event.original.stopPropagation();
        }
    });

    r.observe({
        "mapTowns searchTowns": function() {
            var search = r.get("searchTowns") || [];
            var map = r.get("mapTowns") || [];

            r.set("towns.data", unique(search.concat(map), "key"));

            if (r.get("towns.data.length") < (r.get("towns.page") * r.get("towns.pageSize"))) {
                r.set("towns.page", 1);
            }
        },
    });

    //functions
    function search() {
        var srcOptions = r.get("srcOptions");
        r.set("towns.page", 1);
        r.set("searchTowns", []);

        townSearch(srcOptions)
            .done(function(results) {
                if (results && results.data && results.data.length > 0) {
                    var town = results.data[0];
                    r.set("searchTowns", results.data);
                    selectTown(town);
                } else {
                    r.set("town", null);
                    r.set("searchTowns", []);
                    map.fire("search", srcOptions.search);
                }
                r.set("showResult", true);
            })
            .fail(function(results) {
                r.set("error", err(results).message);
            });
    }

    function selectTown(town) {
        setTheme(town.color);
        map.fire("selectTown", town);
        scrollToResult();
    }

    function nextTowns(from) {
        if (!r.get("srcOptions.search").trim()) {
            return;
        }

        var srcOptions = $.extend({
            from: from
        }, r.get("srcOptions"));

        townSearch(srcOptions)
            .done(function(results) {
                if (results && results.data && results.data.length > 0) {
                    r.set("searchTowns", r.get("searchTowns").concat(result.data));
                }
            })
            .fail(function(results) {
                r.set("error", err(results).message);
            });
    }

    function refreshUser() {
        getUser()
            .done(function(result) {
                r.set("currentUser", result.data);
            })
            .fail(function(result) {
                r.set("error", err(result).message);
            });
    }

    function scrollToResult() {
        $("html, body").animate({
            scrollTop: r.nodes.fixedTrigger.offsetTop + 1
        }, 300);
    }

});
