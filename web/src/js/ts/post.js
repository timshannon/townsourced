// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
import csrf from "./csrf";

export
var maxImages = 10;
export
var maxTitle = 300;

export
var editDuration = 300 * 1000; // seconds in ms

export
var reportReasons = [
    "Post is SPAM, or contains other unsolicited or improper content for this town",
    "Post contains illegal activity, or facilitates illegal activity",
    "Post contains hate speech, harassment, or bullying",
];

export
var categories = {
    "notice": "Notices",
    "buysell": "Buy & Sell",
    "event": "Events",
    "jobs": "Jobs",
    "volunteer": "Volunteer",
    "housing": "Housing",
};

export
var statuses = {
    "draft": "Draft",
    "published": "Published",
    "closed": "Closed",
};



export

function get(id) {
    "use strict";
    return csrf.ajax({
        type: "GET",
        dataType: "json",
        url: "/api/v1/post/" + id,
    });
}

export

function newPost(post, saveAsDraft) {
    "use strict";
    var data = $.extend({
        draft: saveAsDraft,
    }, post);

    return csrf.ajax({
        type: "POST",
        dataType: "json",
        url: "/api/v1/post/",
        data: JSON.stringify(data),
    });
}

export

function save(post) {
    "use strict";

    var data = $.extend({
        draft: true,
    }, post);

    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + post.key,
        data: JSON.stringify(data),
    });

}

export

function publish(post) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + post.key,
        data: JSON.stringify(post),
    });
}

export

function close(postKey, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            close: true,
            vertag: vertag,
        }),
    });

}

export

function reopen(postKey, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            reopen: true,
            vertag: vertag,
        }),
    });

}

export

function unpublish(postKey, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            unpublish: true,
            vertag: vertag,
        }),
    });
}

export

function setNotifyOnComment(postKey, vertag, notifyOnComment) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            notifyOnComment: notifyOnComment,
            vertag: vertag,
        }),
    });
}

export

function moderate(postKey, townKey, reason, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            moderatorTown: townKey,
            moderatorReason: reason,
            vertag: vertag,
        }),
    });

}

export

function removeModeration(postKey, townKey, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            removeModeration: true,
            moderatorTown: townKey,
            vertag: vertag,
        }),
    });

}

export

function isModerated(post, optTownKey) {
    "use strict";
    if (!post || !post.moderation) {
        return false;
    }

    if (!optTownKey) {
        if (post.moderation.length < 1) {
            return false;
        } else {
            return true;
        }
    }

    for (var i = 0; i < post.moderation.length; i++) {
        if (post.moderation[i].town == optTownKey) {
            return true;
        }
    }

    return false;
}

export

function report(postKey, reason, vertag) {
    "use strict";
    return csrf.ajax({
        type: "PUT",
        dataType: "json",
        url: "/api/v1/post/" + postKey,
        data: JSON.stringify({
            report: reason,
            vertag: vertag,
        }),
    });

}

export

function price(post) {
    "use strict";

    if (!post || !post.prices || post.prices.length === 0) {
        return "";
    }
    if (post.prices.length === 1) {
        return "$" + post.prices[0].toFixed(2);
    }

    var prices = post.prices.slice();
    prices.sort(function(a, b) {
        return a - b;
    });

    var first = prices[0].toFixed(2);
    var last = prices[prices.length - 1].toFixed(2);

    if (first == last) {
        return "$" + first;
    }

    return "$" + first + " - " + "$" + last;
}
