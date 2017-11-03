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
        var redirectUrl = buildRedirect();
        new3rdPartyToken("facebook", returnURL || window.location)
            .done(function(tknResult) {
                csrf.ajax({
                        type: "GET",
                        url: "/api/v1/session/facebook/"
                    })
                    .done(function(result) {
                        window.location = "https://www.facebook.com/dialog/oauth/?" +
                            $.param({
                                "client_id": result.data.appID,
                                "redirect_uri": redirectUrl,
                                "response_type": "code",
                                "state": tknResult.data,
                                "scope": "email",
                            });
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
    getSession: function(code) {
        return csrf.ajax({
            type: "GET",
            url: "/api/v1/session/facebook/?" +
                $.param({
                    "code": code,
                    "redirect_uri": buildRedirect(),
                }),
        });
    },
    newUser: function(username, email, userID, userToken, appToken) {
        return csrf.ajax({
            type: "POST",
            url: "/api/v1/session/facebook/",
            data: JSON.stringify({
                username: username,
                email: email,
                userID: userID,
                userToken: userToken,
                appToken: appToken,
            }),
        });

    },
    disconnect: function() {
        return csrf.ajax({
            type: "DELETE",
            url: "/api/v1/session/facebook/",
        });

    },
};

function buildRedirect() {
    return urlJoin(origin(), "3rdparty");
}
