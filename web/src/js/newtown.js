// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Map from "./components/map";
import Alert from "./components/alert";
import Header from "./components/header";

import * as Town from "./ts/town";
import {
    origin, urlify, htmlPayload
}
from "./ts/util";

import {
    err
}
from "./ts/error";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            map: Map,
            alert: Alert,
            header: Header,
        },
        data: function() {
            return {
                "origin": origin(),
                error: null,
                maxTownKey: Town.maxTownKey,
                maxTownDescription: Town.maxTownDescription,
                maxTownName: Town.maxTownName,
                currentUser: htmlPayload("userPayload"),
            };
        },
    });

    //ractive events
    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (!currentUser) {
                r.set("authenticated", false);
                return;
            }
            r.set("authenticated", true);
        },
        "login": function(event) {
            event.original.preventDefault();
            r.findComponent("navbar").fire("login", "Login to post a new town");
        },
        "setTownName": function(event) {
            if (r.get("nameErr") && r.get("name")) {
                r.set("nameErr", null);
            }
            r.set("key", urlify(event.context.name).slice(0, (Town.maxTownKey-1)));
        },
        "setKey": function(event) {
            r.set("key", urlify(r.get("key")));
        },
        "validateDesc": function() {
            if (r.get("descriptionErr") && r.get("description")) {
                r.set("descriptionErr", null);
            }

            if (!r.get("description")) {
                return;
            }

            Town.validateDescription(r.get("description"))
                .fail(function(result) {
                    r.set("descriptionErr", result.message);
                });
        },
        "validateURL": function() {
            r.set("urlErr", null);
            if (!r.get("key")) {
                return;
            }
            Town.validateURL(r.get("key"))
                .fail(function(result) {
                    r.set("urlErr", result.message);
                });
        },
        "createTown": function(event) {
            event.original.preventDefault();
            r.set("urlErr", null);
            r.set("nameErr", null);
            r.set("descriptionErr", null);
            r.set("locationErr", null);

            var name = $.Deferred();
            var description = Town.validateDescription(r.get("description"));
            var location = $.Deferred();
            var url = Town.validateURL(r.get("key"));

            if (r.get("name")) {
                name.resolve();
            } else {
                name.reject();
                r.set("nameErr", "Name is required");
            }

            if (r.get("longitude") && r.get("latitude")) {
                location.resolve();
            } else {
                location.reject();
                r.set("locationErr", "A Location is required");
            }
            url.fail(function(result) {
                r.set("urlErr", result.message);
            });

            description.fail(function(result) {
                r.set("descriptionErr", result.message);
            });

            $.when(url, name, description, location)
                .done(function() {
                    //Post new town
                    //key, name, description, location
                    Town.newTown(r.get("key"), r.get("name"), r.get("description"), r.get("longitude"), r.get("latitude"),
                            r.get("isPrivate"))
                        .done(function(result) {
                            window.location = "/town/" + result.data.key + "/settings/";
                        })
                        .fail(function(result) {
                            r.set("error", err(result).message);
                        });
                });

        },
        "map.search": function() {
            r.set("locationErr", null);
        },
    });

    //functions


});
