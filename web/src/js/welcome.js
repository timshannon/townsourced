// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

//Ractive + components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import TownPanel from "./components/townPanel";
import Modal from "./components/modal";
import Alert from "./components/alert";
import Image from "./components/image";
import ImageUpload from "./components/imageUpload";

import fade from "./lib/ractive-transitions-fade";

//ts libs
import {
    htmlPayload,
    addPager,
}
from "./ts/util";

import {
    err
}
from "./ts/error";
import * as User from "./ts/user";
import {
    join,
    leave,
    setTheme,
    isMember,
}
from "./ts/town";
import csrf from "./ts/csrf";

$(document).ready(function() {
    "use strict";

    // necessary if csrf token isn't in memory yet
    csrf.get(htmlPayload("csrf"));

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            townPanel: TownPanel,
            modal: Modal,
            alert: Alert,
            image: Image,
            imageUpload: ImageUpload,
        },
        transitions: {
            townLoad: function(t) {
                fade(t, {
                    delay: (Math.random() * (600 - 300) + 300),
                });
            },
        },
        data: function() {
            var steps = ["profile"];
            var towns = htmlPayload("towns");
            if (towns && towns.length > 0) {
                steps.unshift("towns");
            }
            return {
                steps: steps,
                stepIndex: 0,
                user: htmlPayload("user"),
                towns: {
                    pageSize: 9,
                    page: 1,
                    data: towns,
                },
                location: htmlPayload("location"),
                town: null,
                error: null,
                isMember: function(town) {
                    var u = this.get("user");
                    if (!town || !u) {
                        return false;
                    }
                    return isMember(town, u);
                },
                nameChange: "",
            };
        },
        onconfig: function() {
            addPager(this);
        },
    });

    //ractive events
    r.on({
        "next": function(event) {
            if (r.get("stepIndex") >= (r.get("steps.length") - 1)) {
                return;
            }

            r.add("stepIndex");
            scrollToTop();
        },
        "prev": function(event) {
            r.set("stepIndex", Math.max(r.get("stepIndex") - 1, 0));
            scrollToTop();
        },
        "finish": function(event) {
            if (!r.get("nameChange")) {
                    window.location = "/";
					return;
			}
            User.setName(r.get("nameChange"), r.get("user.vertag"))
                .done(function() {
                    window.location = "/";
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "join townPanel.join": function(event, town) {
            join(town.key)
                .done(function(result) {
                    refreshUser();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "leave townPanel.leave": function(event, town) {
            leave(town.key)
                .done(function(result) {
                    refreshUser();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "moreInfo": function(event, town) {
            r.set("town", town);
            setTheme(town.color);
            $("#infoPanel").modal();
        },
        "moreTowns": function() {
            r.nextPage("towns");
            $("#moreTowns").blur();
            scrollToTop();
        },
        "backTowns": function() {
            r.prevPage("towns");
            $("#moreTowns").blur();
            scrollToTop();
        },
        "setLocation": function(event) {
            event.original.preventDefault();
            if ("geolocation" in navigator) {
                navigator.geolocation.getCurrentPosition(function(position) {
                    window.location = window.location.pathname + "?" + $.param({
                        latitude: position.coords.latitude,
                        longitude: position.coords.longitude,
                    }, true);

                }, function(error) {
                    r.set("error", "Sorry, we were unable to get your location: " + error.message);
                }, {
                    timeout: 10000,
                });
            } else {
                r.set("error", "We cannot get your current location from this browser, sorry.");
            }


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
                    refreshUser();
                    $("#imageModal").modal("hide");
                })
                .fail(function(result) {
                    r.findComponent("imageUpload").set("error", err(result).message);
                });
        },
        "cancel": function(event) {
            event.original.preventDefault();
        },
        "setName": function(event) {
            if (!r.get("nameChange")) {
                return;
            }
            User.setName(r.get("nameChange"), r.get("user.vertag"))
                .done(function() {
                    refreshUser();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
    });

    function refreshUser() {
        User.get()
            .done(function(result) {
                r.set("user", result.data);
            })
            .fail(function(result) {
                r.set("error", err(result).message);
            });
    }

    function scrollToTop() {
        $("html, body").animate({
            scrollTop: 0,
        }, 300);
    }



});
