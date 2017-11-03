// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
import csrf from "./csrf";

export
var maxLength = 50000;
export
var resultLimit = 40;
export
var editDuration = 60000;
export
var maxDepth = 7;

export

function get(postKey, options) {
    "use strict";
    var query = {};
    options = options || {};

    query.limit = options.limit || resultLimit;
    query.from = options.from || 0;
    query.sort = options.sort;

    return csrf.ajax({
        type: "GET",
        dataType: "json",
        url: "/api/v1/post/" + postKey + "/comment/?" + $.param(query, true),
    });
}

export

function getChildren(postKey, parentKey, options) {
    "use strict";
    var query = {};
    options = options || {};

    query.limit = options.limit || resultLimit;
    query.from = options.from || 0;
    query.sort = options.sort;

    return csrf.ajax({
        type: "GET",
        dataType: "json",
        url: "/api/v1/post/" + postKey + "/comment/" + parentKey + "?" + $.param(query, true),
    });
}

export

function newComment(postKey, comment) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        dataType: "json",
        url: "/api/v1/post/" + postKey + "/comment",
        data: JSON.stringify({
            comment: comment,
        }),
    });

}

export

function reply(postKey, parentKey, comment) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        dataType: "json",
        url: "/api/v1/post/" + postKey + "/comment/" + parentKey,
        data: JSON.stringify({
            comment: comment,
        }),
    });
}
