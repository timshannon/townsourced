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
        var redirectURL = urlJoin(origin(), "3rdparty/");

        $.when(new3rdPartyToken("google", returnURL || window.location),
                csrf.ajax({
                    type: "GET",
                    url: "/api/v1/session/google/"
                }))
            .fail(function(result) {
                window.location = redirectURL + "?" + $.param({
                    "error": true
                });
            })
            .done(function(token, oidCFG) {
                token = token[0].data;
                oidCFG = oidCFG[0].data;
                window.location = oidCFG.authURL + "?" +
                    $.param({
                        "client_id": oidCFG.clientID,
                        "response_type": "code",
                        "redirect_uri": redirectURL,
                        "state": token,
                        "scope": "openid email https://www.googleapis.com/auth/plus.me",
                    });
            });
    },
    getSession: function(code) {
        return csrf.ajax({
            type: "GET",
            url: "/api/v1/session/google/?" +
                $.param({
                    "code": code,
                    "redirect_uri": urlJoin(origin(), "3rdparty/"),
                }),
        });
    },
    newUser: function(username, email, userID, idToken, accessToken) {
        return csrf.ajax({
            type: "POST",
            url: "/api/v1/session/google/",
            data: JSON.stringify({
                username: username,
                email: email,
                userID: userID,
                idToken: idToken,
                accessToken: accessToken,
            }),
        });
    },
    disconnect: function() {
        return csrf.ajax({
            type: "DELETE",
            url: "/api/v1/session/google/",
        });

    },
};
