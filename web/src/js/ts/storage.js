// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

import {
    unique,
	findIndex,
}
from "./util";

import {
    current as currentUser, get as getUser
}
from "./user";
import {
    get as getTown
}
from "./town";

export
// skips parsing
function getStr(key) {
    "use strict";
    return window.localStorage.getItem(key);
}

export

function get(key) {
    "use strict";
    return JSON.parse(window.localStorage.getItem(key));
}


export

function set(key, value) {
    "use strict";
    if (typeof value === 'object') {
        window.localStorage.setItem(key, JSON.stringify(value));
    } else {
        window.localStorage.setItem(key, value);
    }
}

export

function remove(key) {
    "use strict";
    window.localStorage.removeItem(key);
}

export

function push(key, value) {
    "use strict";
    var item = get(key);
    if (!item) {
        item = [];
    }

    if (!Array.isArray(item)) {
        throw "Cannot push data into local storage item, because it's not an array";
    }
    item.push(value);
    set(key, item);
}

export

function pushUniq(key, value, optCompareKey) {
    "use strict";
    var item = get(key);
    if (!item) {
        item = [];
    }
    if (!Array.isArray(item)) {
        throw "Cannot pushUniq data into local storage item, because it's not an array";
    }

    if (Array.isArray(value)) {
        item = item.concat(value);
    } else {
        item.push(value);
    }
    set(key, unique(item, optCompareKey));
}

export

function addTown(town) {
    "use strict";
    var towns = get("towns");
    if (!towns) {
        towns = [];
        set("towns", towns);
    }

    if (towns.indexOf(town) !== -1) {
        return;
    }


    getTown(town)
        .done(function(result) {
            pushUniq("towns", result.data.key);
        });

}

export

function addUser(user) {
    "use strict";
    var users = get("atUsers");
    if (!users) {
        users = [];
        set("atUsers", users);
    }

    if (findIndex(users, user, "username") !== -1) {
        return;
    }

    if (currentUser && user === currentUser.username) {
        pushUniq("atUsers", currentUser);
        return;
    }

    getUser(user)
        .done(function(result) {
            pushUniq("atUsers", result.data);
        });
}
