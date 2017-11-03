// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive + components
import Ractive from "./ts/ractiveInit";
import Login from "./components/login";
import Contact from "./components/contact";

// ts libs
import {
    err
}
from "./ts/error";

//3rd party
import "./lib/bootstrap/modal";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "#loginComponent",
        template: "#tLogin",
        components: {
            login: Login,
        },
    });

    var rc = new Ractive({
        el: "#contactComponent",
        template: "#tContact",
        components: {
            contact: Contact,
        },
    });

    r.on({
        "login": function() {
            r.findComponent("login").fire("doLogin");
        }
    });

    $("#loginBtn").click(function(event) {
        event.preventDefault();
        $("#loginModal").on("shown.bs.modal", function() {
            $("#username").focus();
        });

        $("#loginModal").modal();
    });

    $("#contactBtn, #contactLogin").click(function(event) {
        event.preventDefault();

        $("#contactModal").modal();
    });



});
