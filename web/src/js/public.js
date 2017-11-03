// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */


// Ractive + components
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Image from "./components/image";
import Login from "./components/login";
import Modal from "./components/modal";
import TSIcon from "./components/tsIcon";

// ts libs
import {
    htmlPayload,
    since,
}
from "./ts/util";
import Facebook from "./ts/facebook";
import Google from "./ts/google";
import Twitter from "./ts/twitter";
import {
    login
}
from "./ts/user";


$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            image: Image,
            login: Login,
            modal: Modal,
            tsIcon: TSIcon,
        },
        data: function() {
            var posts = htmlPayload("local") || [];
            var towns = (htmlPayload("towns") || []);

            return {
                local: posts,
				postTowns: towns,
                towns: towns.slice(0, 6 - posts.length),
                since: since,
                waiting: false,
                username: null,
                password: null,
                loginErr: null,
                passwordErr: null,
                postTown: function(post) {
                    var pt = this.get("postTowns");
                    if (!post || !post.townKeys || !pt) {
                        return "";
                    }

                    for (var i = 0; i < post.townKeys.length; i++) {
                        for (var t = 0; t < pt.length; t++) {
                            if (pt[t].key == post.townKeys[i]) {
                                return pt[t].name;
                            }
                        }
                    }

                    return "";
                },
            };
        },
    });

    var loginComp = r.findComponent("login"); // modal login window

    r.on({
        signupLink: function(event) {
            loginComp.fire("reset");
            loginComp.set("signup", true);

            $("#loginModal").on("shown.bs.modal", function() {
                $("#username").focus();
            });

            $("#loginModal").modal();
        },
        forgotPasswordLink: function(event) {
            event.original.preventDefault();
            loginComp.fire("reset");
            loginComp.set("forgotPass", true);

            $("#loginModal").on("shown.bs.modal", function() {
                $("#username").focus();
            });

            $("#loginModal").modal();
        },
        townSearch: function(event) {
            event.original.preventDefault();
            window.location = "/town/?search=" + encodeURIComponent(r.get("townSearch"));
        },
        loginFacebook: function() {
            Facebook.startLogin("/");
        },
        loginGoogle: function() {
            Google.startLogin("/");
        },
        loginTwitter: function() {
            Twitter.startLogin("/");
        },
        usernameBlur: function() {
            r.set("loginErr", null);
        },
        passwordBlur: function() {
            r.set("passwordErr", null);
        },
        doLogin: function(event) {
            if (event) {
                event.original.preventDefault();
            }
            r.set("waiting", true);
            r.set("loginErr", null);
            r.set("passwordErr", null);

            if (!r.get("username")) {
                r.set("loginErr", "You must provide a username or email address");
                r.set("waiting", false);
                return;
            }

            if (!r.get("password")) {
                r.set("passwordErr", "You must provide a password");
                r.set("waiting", false);
                return;
            }

            login(r.get("username"), r.get("password"), r.get("rememberMe"))
                .done(function(result) {
                    window.location.reload();
                })
                .fail(function(result) {
                    result = result.responseJSON;
                    r.set("loginErr", result.message);
                    r.set("waiting", false);
                });
        },

    });
});
