// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
import csrf from "./csrf";
import {
    loadAndStore
}
from "./util";
import "../lib/tinycolor";
import {
    current as currentUser
}
from "./user";

export
var announcementTown = "announcements";

export
var maxTownDescription = 500;

export
var maxTownName = 200;

export
var maxTownKey = 127;

export
var defaultTownColor = "#246b08";

export
var headerHeight = 500;
export
var headerWidth = 1950;

export

function get(town) {
    "use strict";
    return csrf.ajax({
        type: "GET",
        dataType: "json",
        url: "/api/v1/town/" + town + "/",
    });
}
export

function newTown(key, name, description, longitude, latitude, isPrivate) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        dataType: "json",
        url: "/api/v1/town/",
        data: JSON.stringify({
            key: key,
            name: name,
            description: description,
            longitude: longitude,
            latitude: latitude,
            private: isPrivate,
        }),
    });
}

export

function validateDescription(description) {
    "use strict";
    var dfr = $.Deferred();

    if (description.trim() === "") {
        dfr.reject({
            message: "You must add a town description"
        });
        return dfr;
    }

    if (description.length > maxTownDescription) {
        dfr.reject({
            message: "A town description must be less than " + maxTownDescription + " characters.",
        });
        return dfr;
    }

    dfr.resolve();

    return dfr;
}

export

function validateURL(key) {
    "use strict";
    var dfr = $.Deferred();

    if (key === "") {
        dfr.reject({
            message: "You must specify a url"
        });
        return dfr;
    }

    if (!isTownKey(key)) {
        dfr.reject({
            message: "Town URLs can only contain letters, numbers, and dashes."
        });
        return dfr;
    }

    if (key.length > maxTownKey) {
        dfr.reject({
            message: "A town url must be less than " + maxTownKey + " characters.",
        });
        return dfr;
    }

    get(key)
    .done(function() {
            dfr.reject({
                message: "URL already exists, please choose another."
            });
        })
        .fail(function() {
            dfr.resolve();
        });
    return dfr;
}
export

function isModerator(town, user) {
    "use strict";
    if (!user) {
        return false;
    }

    if (user.admin) {
        return true;
    }

    if (!town || !town.moderators) {
        return false;
    }

    for (var i = 0; i < town.moderators.length; i++) {
        var mod = town.moderators[i];
        if (mod.username == user.username) {
            mod.startDate = Date.parse(mod.start);
            mod.endDate = Date.parse(mod.end);
            if (mod.startDate && mod.startDate < Date.now() && mod.startDate > 0) {
                if (mod.endDate && mod.endDate > 0 && mod.endDate < Date.now()) {
                    return false;
                }
                return true;
            }
        }
    }
    return false;
}

export

function isInvitedMod(town, user) {
    "use strict";
    if (!town.moderators) {
        return false;
    }

    for (var i = 0; i < town.moderators.length; i++) {
        var mod = town.moderators[i];
        if (mod.username == user.username) {
            mod.inviteDate = Date.parse(mod.inviteSent);
            if (mod.inviteDate && mod.inviteDate > 0 && mod.inviteDate < Date.now()) {
                return true;
            } else {
                return false;
            }
        }
    }
    return false;
}
export

function setHeaderImage(townKey, imageKey, x0, y0, x1, y1, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/image/",
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

function removeHeaderImage(townKey, vertag) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/image/",
        data: JSON.stringify({
            vertag: vertag,
        }),
    });
}
export

function setName(townKey, newName, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/",
        data: JSON.stringify({
            name: newName,
            vertag: vertag,
        }),
    });
}
export

function setDescription(townKey, newDescription, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/",
        data: JSON.stringify({
            description: newDescription,
            vertag: vertag,
        }),
    });
}

export

function setInformation(townKey, newInformation, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/",
        data: JSON.stringify({
            information: newInformation,
            vertag: vertag,
        }),
    });
}

export

function setColor(townKey, newColor, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/",
        data: JSON.stringify({
            color: newColor,
            vertag: vertag,
        }),
    });
}

export

function setPrivacy(townKey, isPrivate, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/",
        data: JSON.stringify({
            private: isPrivate,
            vertag: vertag,
        }),
    });
}

export

function setTheme(color) {
    "use strict";
    $("#townTheme").remove();

    color = color || defaultTownColor;
    var tc = tinycolor(color);
    var style =
        ".town-btn {" + buttonStyle(tc) + "} " +
        ".town-btn:hover, .town-btn:focus, .town-btn:active, " +
        ".town-btn.active, .open > .dropdown-toggle.town-btn {" + buttonStyle(tc, true) + "} " +
        ".town-background {" + backgroundStyle(tc) + "} " +
        ".town-border {border-color: " + color + "} " +
        ".town-text {" + textStyle(tc) + "} ";
    $("<style id='townTheme'>" + style + "</style>").appendTo("head");
}
export

function inviteMod(townKey, newMod) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/town/" + townKey + "/moderator/",
        data: JSON.stringify({
            moderator: newMod,
        }),
    });

}

export

function acceptMod(townKey) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/moderator/",
    });
}

export

function removeMod(townKey) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/moderator/",
    });
}

export

function isTownKey(key) {
    "use strict";
    return /^[a-zA-Z0-9-]*$/i.test(key);
}

export

function removeInvite(townKey, username) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/invite/",
        data: JSON.stringify({
            invitee: username,
        }),
    });
}

export

function invite(townKey, username, email) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/town/" + townKey + "/invite/",
        data: JSON.stringify({
            invitee: username,
            email: email,
        }),
    });
}

export

function requestInvite(townKey) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/town/" + townKey + "/invite/request/",
    });
}

export

function acceptInviteRequest(townKey, invitee) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/invite/request/",
        data: JSON.stringify({
            invitee: invitee,
        }),
    });
}

export

function rejectInviteRequest(townKey, invitee) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/invite/request/",
        data: JSON.stringify({
            invitee: invitee,
        }),
    });
}

export

function invited(town, user) {
    "use strict";
    if (!town.private) {
        return true;
    }
    if (isModerator(town, user)) {
        return true;
    }

    if (!town.invites) {
        return false;
    }

    for (var i = 0; i < town.invites.length; i++) {
        if (town.invites[i] === user.username) {
            return true;
        }
    }

    return false;
}

export

function isMember(town, user) {
    "use strict";
    if (!user || !user.townKeys) {
        return false;
    }

    for (var i = 0; i < user.townKeys.length; i++) {
        if (user.townKeys[i].key === town.key) {
            return true;
        }
    }
    return false;
}

export

function join(townKey) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/user/me/town/" + townKey,
    });
}

export

function leave(townKey) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/user/me/town/" + townKey,
    });
}

export

function addAutoModCategory(townKey, category) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/town/" + townKey + "/automod/category",
        data: JSON.stringify({
            category: category,
        }),
    });
}

export

function removeAutoModCategory(townKey, category) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/automod/category",
        data: JSON.stringify({
            category: category,
        }),
    });
}

export

function setAutoModMinUserDays(townKey, minUserDays) {
    "use strict";

    var dfr = $.Deferred();
    if (!minUserDays) {
        minUserDays = 0;
    }
    if (!$.isNumeric(minUserDays)) {
        dfr.reject("Invalid number");
        return dfr;
    }
    if (minUserDays < 0) {
        dfr.reject("Number of days must be positive");
        return dfr;
    }

    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/automod",
        data: JSON.stringify({
            minUserDays: minUserDays,
        }),
    });
}
export

function setAutoModMaxNumLinks(townKey, maxNumLinks) {
    "use strict";

    var dfr = $.Deferred();
    if (!maxNumLinks) {
        maxNumLinks = 0;
    }
    if (!$.isNumeric(maxNumLinks)) {
        dfr.reject("Invalid number");
        return dfr;
    }
    if (maxNumLinks < 0) {
        dfr.reject("Number of links must be positive");
        return dfr;
    }

    return csrf.ajax({
        type: "PUT",
        url: "/api/v1/town/" + townKey + "/automod",
        data: JSON.stringify({
            maxNumLinks: maxNumLinks,
        }),
    });
}

export

function addAutoModUser(townKey, username) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/town/" + townKey + "/automod/user",
        data: JSON.stringify({
            user: username,
        }),
    });
}

export

function removeAutoModUser(townKey, username) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/automod/user",
        data: JSON.stringify({
            user: username,
        }),
    });
}

export

function addAutoModRegexp(townKey, regexp, reason) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/town/" + townKey + "/automod/regexp",
        data: JSON.stringify({
            regexp: regexp,
            reason: reason,
        }),
    });
}

export

function removeAutoModRegexp(townKey, regexp) {
    "use strict";
    return csrf.ajax({
        type: "DELETE",
        url: "/api/v1/town/" + townKey + "/automod/regexp",
        data: JSON.stringify({
            regexp: regexp,
        }),
    });
}

export

function canPost(town, optUser) {
    "use strict";
    if (!town) {
        return false;
    }

    if (!town.private && town.key != announcementTown) {
        return true;
    }

    if (!optUser) {
        if (!currentUser) {
            return false;
        }
        optUser = currentUser;
    } else if (typeof optUser != "string") {
        optUser = optUser;
    }

    if (isModerator(town, optUser)) {
        return true;
    }

    if (!town.invites) {
        return false;
    }

    for (var i = 0; i < town.invites.length; i++) {
        if (optUser.username === town.invites[i]) {
            return true;
        }
    }

    return false;
}

export

function canJoin(town, optUser) {
    "use strict";
    if (!town) {
        return false;
    }
    if (!town.private) {
        return true;
    }

    if (!optUser) {
        if (!currentUser) {
            return false;
        }
        optUser = currentUser;
    } else if (typeof optUser != "string") {
        optUser = optUser;
    }

	if(isModerator(town, optUser)) {
		// mods are auto invited
		return true;
	}


    if (!town.invites) {
        return false;
    }

    for (var i = 0; i < town.invites.length; i++) {
        if (optUser.username === town.invites[i]) {
            return true;
        }
    }

    return false;
}


export
// loadAndStoreTown loads the passed in townkeys into either optKeypath, or townLoad
// if optAltSource is provided (array or object), it'll check there for existing town data
function loadAndStoreTown(ractive, townKeys, optKeypath, optAltSource) {
    "use strict";
    var kp = optKeypath || "townLoad";

    loadAndStore(ractive, townKeys, kp, get, optAltSource);
}

export

function search(options) {
    "use strict";

    $.each(options, function(k, v) {
        if (!v) {
            delete options[k];
        }
    });

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/town/?" + $.param(options, true),
    });
}


//theme
var townTextColors = ["#fff", "#333"];

function buttonStyle(color, hover) {
    "use strict";
    var style = "";
    if (hover) {
        style += "background-color: " + tinycolor(color.toString()).darken(10) + ";" +
            "border-color: " + tinycolor(color.toString()).darken(10) + ";";

    } else {
        style += "background-color: " + color + ";" +
            "border-color: " + color + ";";
    }

    style += "color: " + tinycolor.mostReadable(tinycolor(color.toString()).darken(10), townTextColors, {
        includeFallbackColors: false,
        level: "AAA",
        size: "large"
    }).toHexString() + ";";
    return style;
}

function textStyle(color) {
    "use strict";
    return "color: " + color + ";";
}

function backgroundStyle(color) {
    "use strict";
    return "background-color: " + color + ";" +
        "color: " + tinycolor.mostReadable(tinycolor(color.toString()).darken(10), townTextColors, {
            includeFallbackColors: false,
            level: "AAA",
            size: "large"
        }).toHexString() + ";";
}
