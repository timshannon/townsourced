// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
import csrf from "./csrf";
import {
    loadAndStore,
    findIndex,
    isEmail,
}
from "./util";

export
var current = {};

export
var maxUsername = 127;

export

function setCurrent(user) {
    "use strict";
    current = user;
}

export

function
get(userid) {
    "use strict";
    if (!userid) {
        userid = "me";
    }
    var g = csrf.ajax({
        type: "GET",
        url: "/api/v1/user/" + userid + "/",
    });
    if (userid === "me") {
        g.done(function(result) {
            current = result.data;
        });
    }
    return g;
}

export

function match(matchStr, optLimit) {
    "use strict";
    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/?" +
            $.param({
                "limit": optLimit || 10,
                "match": matchStr,
            }),
    });

}

export

function signup(username, email, password) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/user/",
        data: JSON.stringify({
            username: username,
            email: email,
            password: password,
        }),
    });
}
export

function login(username, password, rememberMe) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/session/",
        data: JSON.stringify({
            username: username,
            password: password,
            rememberMe: rememberMe,
        }),
    });
}
export

function logout() {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/session/",
    });
}
export

function emailCheck(email) {
    "use strict";
    return emailCheck(email);
}
export

function validateNew(username) {
    "use strict";
    var dfr = $.Deferred();

    if (username === "") {
        dfr.reject({
            message: "You must specify a username"
        });
        return dfr;
    }

    if (username.length > maxUsername) {
        dfr.reject({
            message: "Usernames must be less than 127 characters."
        });
        return dfr;
    }


    if (!isUsername(username)) {
        dfr.reject({
            message: "Usernames can only contain letters, numbers, and dashes."
        });
        return dfr;
    }

    get(username)
    .done(function() {
            dfr.reject({
                message: "Username already exists, please choose another."
            });
        })
        .fail(function() {
            dfr.resolve();
        });
    return dfr;
}
export

function validateEmail(email) {
    "use strict";
    var dfr = $.Deferred();

    if (email === "") {
        dfr.reject({
            message: "You must specify an email address"
        });
        return dfr;
    }

    if (!isEmail(email)) {
        dfr.reject({
            message: "Invalid email address"
        });
        return dfr;
    }

    emailCheck(email)
        .done(function() {
            dfr.resolve();
        })
        .fail(function(result) {
            dfr.reject({
                message: result.responseJSON.message,
            });
        });
    return dfr;
}
export

function new3rdPartyToken(provider, returnURL) {
    "use strict";
    returnURL = returnURL.toString();
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/session/3rdparty/",
        data: JSON.stringify({
            returnURL: returnURL,
            provider: provider,
        }),
    });
}
export

function get3rdPartyState(token) {
    "use strict";
    return csrf.ajax({
        type: "GET",
        url: "/api/v1/session/3rdparty/?token=" + token,
    });
}
export

function setName(newName, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/",
        data: JSON.stringify({
            vertag: vertag,
            name: newName,
        }),
    });
}
export

function setEmail(newEmail, vertag, password) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/",
        data: JSON.stringify({
            vertag: vertag,
            email: newEmail,
            password: password,
        }),
    });
}
export

function setPassword(currentPass, newPassword, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/",
        data: JSON.stringify({
            vertag: vertag,
            password: currentPass,
            newPassword: newPassword,
        }),
    });
}
export

function setProfileImage(imageKey, x0, y0, x1, y1, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/image/",
        data: JSON.stringify({
            imageKey: imageKey,
            x0: x0,
            y0: y0,
            x1: x1,
            y1: y1,
            vertag: vertag,
        }),
    });
}


export

function set(data) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/",
        data: JSON.stringify(data),
    });
}

export

function unreadNotificationCount() {
    "use strict";
    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/me/notifications/?count",
    });
}
export

function unreadNotifications(since, limit) {
    "use strict";
    limit = limit || 100;

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/me/notifications/?" +
            $.param({
                "since": since,
                "limit": limit,
            }),
    });
}

export

function getNotification(key) {
    "use strict";
    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/me/notifications/?" +
            $.param({
                "key": key,
            }),
    });
}

export

function allNotifications(since, limit) {
    "use strict";
    limit = limit || 100;

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/me/notifications/?" +
            $.param({
                "since": since,
                "all": "",
                "limit": limit,
            }),
    });
}

export

function sentNotifications(since, limit) {
    "use strict";
    limit = limit || 100;

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/me/notifications/?" +
            $.param({
                "since": since,
                "sent": true,
                "limit": limit,
            }),
    });
}

export

function markAllNotificationsRead() {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/notifications/",
    });

}
export

function markNotificationRead(key) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/notifications/",
        data: JSON.stringify({
            key: key,
        }),
    });
}
export

function towns(user) {
    "use strict";
	user = user || "me";

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/" + user + "/town/",
    });
}
export

function sendMessage(to, subject, message) {
    "use strict";
    if (!subject || !message) {
        var dfr = $.Deferred();
        dfr.reject("You must specify a subject and message");
        return dfr;
    }

    return csrf.ajax({
        type: "POST",
        url: "/api/v1/user/" + to + "/notifications/",
        data: JSON.stringify({
            subject: subject,
            message: message,
        }),
    });
}

export

function posts(user, optStatus, optSince, optLimit) {
    "use strict";
    optLimit = optLimit || 50;

    if (optStatus == "all") {
        optStatus = "";
    }

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/" + user + "/post/?" +
            $.param({
                "status": optStatus,
                "since": optSince,
                "limit": optLimit,
            }),
    });
}

export

function savedPosts(user, optStatus, optFrom, optLimit) {
    "use strict";
    optLimit = optLimit || 50;

    if (optStatus == "all") {
        optStatus = "";
    }

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/" + user + "/post/saved?" +
            $.param({
                "status": optStatus,
                "from": optFrom,
                "limit": optLimit,
            }),
    });
}

export

function comments(user, options) {
    "use strict";
    var query = {};
    options = options || {};

    query.limit = options.limit || resultLimit;
    query.since = options.since;

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/user/" + user + "/comment/?" +
            $.param(query),
    });
}




export

function loadAndStoreUser(ractive, usernames, optKeypath, optAltSource) {
    "use strict";
    var kp = optKeypath || "userLoad";

    loadAndStore(ractive, usernames, kp, get, optAltSource);
}


export

function savePost(postKey) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/user/me/post/saved/" + postKey,
    });
}

export

function removeSavedPost(postKey) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/user/me/post/saved/" + postKey,
    });
}

export

function isSavedPost(user, postKey) {
    "use strict";
    if (!user || !user.savedPosts) {
        return false;
    }
    return findIndex(user.savedPosts, postKey, 'key') !== -1;
}


export

function requestPasswordReset(usernameOrEmail) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/forgotpassword",
        data: JSON.stringify({
            username: usernameOrEmail,
        }),
    });
}

export

function resetPassword(token, newPassword) {
    "use strict";

    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/forgotpassword/" + token,
        data: JSON.stringify({
            newPassword: newPassword,
        }),
    });
}

export

function confirmEmail() {
    "use strict";

    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/user/me/confirmemail/",
    });
}


export
function listName(user) {
    "use strict";
	if (!user) {
		return "";
	}

    if (typeof user === "string") {
        return user;
    }

    if (user.name) {
        return user.name;
    }

    if (user.username) {
        return user.username;
    }

    //shouldn't happen
    return "";
}



//not exported
function isUsername(username) {
    "use strict";
    return /^[a-zA-Z0-9-]*$/i.test(username);
}


function emailCheck(email) {
    "use strict";
    return csrf.ajax({
        type: "GET",
        url: "/api/v1/email/?" +
            $.param({
                "email": email,
            }),
    });
}
