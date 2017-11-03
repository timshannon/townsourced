// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

import Ractive from "./ts/ractiveInit";
import Facebook from "./ts/facebook";
import Google from "./ts/google";
import Twitter from "./ts/twitter";

import Spinner from "./components/spinner";
import {
    origin, urlQuery
}
from "./ts/util";

import {
    err
}
from "./ts/error";

import {
    get3rdPartyState, validateEmail as userValidateEmail, validateNew
}
from "./ts/user";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "#main",
        template: "#tMain",
        components: {
            spinner: Spinner,
        },
    });

    var query = urlQuery();
    if (query.error) {
        r.set("returnURL", origin());
        r.set("error", query.error);
        return;
    }

    if (query.state) {
        get3rdPartyState(query.state)
            .done(function(result) {
                r.set("provider", result.data.provider);
                r.set("returnURL", result.data.returnURL);
                if (result.data.provider == "facebook") {
                    facebook(query.code);
                } else if (result.data.provider == "google") {
                    google(query.code);
                } else if (result.data.provider == "twitter") {
                    twitter(query.oauth_verifier, query.state);
                } else {
                    r.set("error", true);
                }
            })
            .fail(function() {
                r.set("error", true);
            });
    } else {
        r.set("returnURL", origin());
        r.set("error", query.error);
        return;
    }

    r.on({
        "usernameBlur": function(event) {
            validateUsername();
        },
        "emailBlur": function(event) {
            validateEmail();
        },
        "continue": function(event) {
            event.original.preventDefault();
            r.set("emailErr", null);
            r.set("usernameErr", null);

            var emlVal = userValidateEmail(event.context.email);
            var usrVal = validateNew(event.context.username);

            emlVal.fail(function(result) {
                r.set("emailErr", result.message);
            });

            usrVal.fail(function(result) {
                r.set("usernameErr", result.message);
            });

            $.when(emlVal, usrVal)
                .done(function() {
                    if (r.get("provider") == "facebook") {
                        Facebook.newUser(event.context.username, event.context.email, event.context.userID,
                                event.context.userToken, event.context.appToken)
                            .done(function() {
								window.location = "/welcome";
                                return;
                            })
                            .fail(function(result) {
                                r.set("usernameErr", err(result).message);
                            });
                    } else if (r.get("provider") == "google") {
                        Google.newUser(event.context.username, event.context.email, event.context.userID,
                                event.context.idToken, event.context.accessToken)
                            .done(function() {
								window.location = "/welcome";
                                return;
                            })
                            .fail(function(result) {
                                r.set("usernameErr", err(result).message);
                            });
                    } else if (r.get("provider") == "twitter") {
                        Twitter.newUser(event.context.username, event.context.email, event.context.token)
                            .done(function() {
								window.location = "/welcome";
                                return;
                            })
                            .fail(function(result) {
                                r.set("usernameErr", err(result).message);
                            });
                    }

                });
        }
    });

    function validateUsername() {
        r.set("usernameErr", null);
        validateNew(r.get("username"))
            .fail(function(result) {
                r.set("usernameErr", result.message);
            });
    }

    function validateEmail() {
        r.set("emailErr", null);
        userValidateEmail(r.get("email"))
            .fail(function(result) {
                r.set("emailErr", result.message);
            });
    }




    // providers
    function facebook(code) {
        Facebook.getSession(code)
            .done(function(result) {
                window.location = r.get("returnURL");
                return;
            })
            .fail(function(result) {
                if (result.responseJSON && result.responseJSON.message == "Username needed") {
                    result = result.responseJSON;
                    r.set("usernameNeeded", true);
                    r.set("username", result.data.username);
                    r.set("email", result.data.email);
                    r.set("userID", result.data.userID);
                    r.set("appToken", result.data.appToken);
                    r.set("userToken", result.data.userToken);
                    validateUsername();
                    validateEmail();
                    $("#username").focus();
                } else {
                    handleErr(result);
                }
            });
    }

    function google(code) {
        Google.getSession(code)
            .done(function(result) {
                window.location = r.get("returnURL");
                return;
            })
            .fail(function(result) {
                if (result.responseJSON && result.responseJSON.message == "Username needed") {
                    result = result.responseJSON;
                    r.set("usernameNeeded", true);
                    r.set("username", result.data.username);
                    r.set("email", result.data.email);
                    r.set("userID", result.data.userID);
                    r.set("idToken", result.data.idToken);
                    r.set("accessToken", result.data.accessToken);
                    validateUsername();
                    validateEmail();
                    $("#username").focus();
                } else {
                    handleErr(result);
                }
            });
    }

    function twitter(code, token) {
        Twitter.getSession(token, code)
            .done(function(result) {
                window.location = r.get("returnURL");
                return;
            })
            .fail(function(result) {
                if (result.responseJSON && result.responseJSON.message == "Username needed") {
                    result = result.responseJSON;
                    r.set("usernameNeeded", true);
                    r.set("username", result.data.username);
                    r.set("email", result.data.email);
                    r.set("userID", result.data.userID);
                    r.set("token", result.data.token);
                    validateUsername();
                    validateEmail();
                    $("#username").focus();
                } else {
                    handleErr(result);
                }
            });
    }

    function handleErr(result) {
        if (result.responseJSON && result.responseJSON.status == "fail") {
            r.set("error", {
                message: result.responseJSON.message
            });
        } else {
            r.set("error", true);
        }
    }
});
