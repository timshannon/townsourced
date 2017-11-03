// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
import * as User from "../../ts/user";
import Facebook from "../../ts/facebook";
import Google from "../../ts/google";
import Twitter from "../../ts/twitter";
import {
    err
}
from "../../ts/error";

import {
    scale
}
from "../../lib/ractive-transition-scale";
export
default {
    isolated: true,
    transitions: {
        scale: scale,
    },
    data: {
        email: "",
        username: "",
        password: "",
        rememberMe: false,
        signup: false,
        forgotPass: false,
        resetSent: false,
        redirect: null,
    },
    oncomplete: function() {
            "use strict";
            var r = this;

            r.on({
                toggleSignup: function(event) {
                    event.original.preventDefault();
                    r.toggle("signup");
                    r.set("loginErr", null);
                    r.set("passwordErr", null);
                    $("#username").focus();
                },
                toggleForgotPassword: function(event) {
                    event.original.preventDefault();
                    r.toggle("forgotPass");
                    r.set("loginErr", null);
                    r.set("passwordErr", null);
                    r.set("resetSent", false);
                    $("#username").focus();
                },
                resetPassword: function(event) {
                    event.original.preventDefault();
                    r.set("waiting", true);
                    r.set("loginErr", null);

                    if (r.get("username") === "") {
                        r.set("loginErr", "You must provide a username or email address");
                        r.set("waiting", false);
                        return;
                    }

                    User.requestPasswordReset(r.get("username"))
                        .done(function(result) {
                            r.set("resetSent", true);
                            r.set("username", "");
                        })
                        .fail(function(result) {
                            r.set("loginErr", err(result).message);
                        })
                        .always(function() {
                            r.set("waiting", false);
                        });

                },
                reset: function() {
                    r.set({
                        "loginErr": null,
                        "passwordErr": null,
                        "emailErr": null,
                        "username": "",
                        "email": "",
                        "password": "",
                        "password2": "",
                        "password2Err": null,
                        "rememberMe": false,
                        "signup": false,
                        "forgotPass": false,
                        "resetSent": false,
                    });
                },
                signup: function(event) {
                    event.original.preventDefault();
                    r.set("waiting", true);
                    r.set("loginErr", null);
                    r.set("passwordErr", null);
                    r.set("emailErr", null);
                    r.set("password2Err", null);

                    validatePassword();
                    var usrVal = User.validateNew(event.context.username);
                    var emlVal = User.validateEmail(event.context.email);
                    var passVal = $.Deferred();

                    if (event.context.password === "") {
                        r.set("passwordErr", "You must provide a password");
                    }

                    if (r.get("password2Err")) {
                        passVal.reject();
                    } else {
                        passVal.resolve();
                    }

                    usrVal.fail(function(result) {
                        r.set("loginErr", result.message);
                    });
                    emlVal.fail(function(result) {
                        r.set("emailErr", result.message);
                    });

                    $.when(usrVal, emlVal, passVal)
                        .fail(function() {
                            r.set("waiting", false);
                        })
                        .done(function() {
                            User.signup(event.context.username, event.context.email, event.context.password)
                                .done(function(result) {
                                    window.location = "/welcome";
                                })
                                .fail(function(result) {
                                    var error = err(result).message;
                                    r.set("waiting", false);
                                    r.set("loginErr", null);
                                    r.set("passwordErr", null);
                                    r.set("emailErr", null);
                                    r.set("password2Err", null);
                                    if (error.indexOf("Invalid password") === -1) {
                                        r.set("loginErr", error);
                                    } else {
                                        r.set("passwordErr", error);
                                    }
                                });
                        });
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

                    User.login(r.get("username"), r.get("password"), r.get("rememberMe"))
                        .done(function(result) {
                            if (r.get("redirect")) {
                                window.location = r.get("redirect");
                            } else {
                                window.location.reload();
                            }
                        })
                        .fail(function(result) {
                            result = result.responseJSON;
                            r.set("loginErr", result.message);
                            r.set("waiting", false);
                        });
                },
                loginFacebook: function() {
                    Facebook.startLogin(r.get("redirect"));
                },
                loginGoogle: function() {
                    Google.startLogin(r.get("redirect"));
                },
                loginTwitter: function() {
                    Twitter.startLogin(r.get("redirect"));
                },
                usernameBlur: function() {
                    r.set("loginErr", null);
                    if (!r.get("username")) {
                        return;
                    }

                    if (!r.get("signup")) {
                        return;
                    }

                    User.validateNew(r.get("username"))
                        .fail(function(result) {
                            r.set("loginErr", result.message);
                        });
                },
                passwordBlur: function() {
                    r.set("passwordErr", null);
                },
                emailBlur: function() {
                    r.set("emailErr", null);
                    if (!r.get("email")) {
                        return;
                    }
                    if (!r.get("signup")) {
                        return;
                    }

                    User.validateEmail(r.get("email"))
                        .fail(function(result) {
                            r.set("emailErr", result.message);
                        });
                },
                password2Blur: function() {
                    validatePassword();
                },
            });

            function validatePassword() {
                r.set("password2Err", null);
                if (!r.get("signup")) {
                    return;
                }

                if (r.get("password2") !== r.get("password")) {
                    r.set("password2Err", "Password does not match!");
                    return;
                }
            }

        } //onrender
};
