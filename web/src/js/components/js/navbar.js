// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
import "../../lib/bootstrap/dropdown";
import * as User from "../../ts/user";
import {
    err
}
from "../../ts/error";

export
default {
    isolated: true,
    transitions: {},
    data: {
        authenticated: false,
        usrLoaded: false,
        title: null,
        town: null,
        altBrand: false,
        notifications: -1,
    },
    oncomplete: function() {
            "use strict";
            var r = this;

            if (!r.get("user")) {
                User.get()
                    .done(function(result) {
                        loadUser(result.data);
                    })
                    .fail(function(result) {
                        var error = err(result);
                        if (error) {
                            r.set("authenticated", false);
                            loadUser(null);
                        }
                    });
            } else {
                User.setCurrent(r.get("user"));
                loadUser(r.get("user"));
            }


            r.on({
                logout: function(event) {
                    if (event) {
                        event.original.preventDefault();
                    }
                    User.logout()
                        .done(function() {
                            window.location = "/";
                        })
                        .fail(function(result) {
                            err(result);
                        });
                },
                login: function(title, redirect) {
                    r.set({
                        "title": title,
                        "redirect": redirect,
                    });
                    r.fire("loginModal");
                },
				loginModal: function(event) {
                    if (event) {
                        event.original.preventDefault();
                    }
                    $("#loginModal").on("shown.bs.modal", function() {
                        $("#username").focus();
                    });

                    if (event) {
                        //called from component
                        r.set({
                            "title": null,
                            "redirect": null,
                        });
                    }

                    r.findComponent("login").fire("reset");

                    $("#loginModal").modal();
                },
                checkNotifications: function() {
                    User.unreadNotificationCount()
                        .done(function(result) {
                            r.set("notifications", result.data);
                        });
                },
            });

            function checkNotifications() {
                User.unreadNotificationCount()
                    .done(function(result) {
                        r.set("notifications", result.data);

                        window.setTimeout(checkNotifications, 30000);
                    });
            }

            function loadUser(currentUser) {
                r.set("user", currentUser);
                r.set("authenticated", (currentUser !== null));
                r.set("usrLoaded", true);
                if (currentUser) {
                    checkNotifications();
                }
                r.fire("userLoaded", r.get("user"));
            }

        } //onrender
};
