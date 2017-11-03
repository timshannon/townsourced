// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive + components
import Ractive from "./ts/ractiveInit";
import UserIcon from "./components/userIcon";

// ts libs
import {
    err
}
from "./ts/error";

import {
    htmlPayload
}
from "./ts/util";

import {
    resetPassword
}
from "./ts/user";

//3rd party


$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            userIcon: UserIcon,
        },
        data: function() {
            return {
                user: htmlPayload("userPayload"),
                token: htmlPayload("tokenPayload"),
                waiting: false,
            };
        },
        oncomplete: function() {
            $("#password").focus();
        },
    });


    r.on({
        "password2Blur": function(event) {
            r.set("password2Err", null);

            if (r.get("password2") !== r.get("password")) {
                r.set("password2Err", "Password does not match!");
                return;
            }
        },
        "setPassword": function(event) {
            if (event) {
                event.original.preventDefault();
            }

            if (r.get("waiting")) {
                return;
            }
            r.set("password2Err", null);
            r.set("passwordErr", null);

            if (!r.get("password")) {
                r.set("passwordErr", "You must specify a password");
                return;
            }

            if (r.get("password2") !== r.get("password")) {
                r.set("password2Err", "Password does not match!");
                return;
            }
            r.set("waiting", true);
            resetPassword(r.get("token"), r.get("password"))
                .done(function() {
                    window.location = "/";
                })
                .fail(function(result) {
                    r.set("passwordErr", err(result).message);
                    r.set("waiting", false);
                });
        }
    });

});
