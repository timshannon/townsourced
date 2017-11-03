// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true */
import csrf from "./csrf";
import {
    origin,
    urlJoin
}
from "./util";

import {
    new3rdPartyToken
}
from "./user";


export
default {
    startLogin: function(returnURL) {
        //Get oauth request token
        var redirectURL = urlJoin(origin(), "3rdparty");
        new3rdPartyToken("twitter", returnURL || window.location)
            .done(function(result) {
                csrf.ajax({
                        type: "POST",
                        url: "/api/v1/session/twitter/",
                        data: JSON.stringify({
                            token: result.data,
                        }),
                    })
                    .done(function(result) {
                        window.location = result.data;
                    })
                    .fail(function() {
window.location = redirectURL + "?" + $.param({
                    "error": true
                });
                    });
            })
            .fail(function() {
                window.location = redirectURL + "?" + $.param({
                    "error": true
                });
            });
    },
    getSession: function(token, code) {
        return csrf.ajax({
            type: "GET",
            url: "/api/v1/session/twitter/?" +
                $.param({
                    "code": code,
                    "token": token,
                }),
        });
    },
    newUser: function(username, email, token) {
        return csrf.ajax({
            type: "POST",
            url: "/api/v1/session/twitter/",
            data: JSON.stringify({
                username: username,
                email: email,
                token: token,
            }),
        });
    },
    disconnect: function() {
        return csrf.ajax({
            type: "DELETE",
            url: "/api/v1/session/twitter/",
        });

    },
};
