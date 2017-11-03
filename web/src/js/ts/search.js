// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
import csrf from "./csrf";

import {
    urlQuery,
}
from "../ts/util";

export
var resultLimit = 40;

export
var sortOptions = {
    none: "Relevance",
    new: "Newest",
    old: "Oldest",
    pricelow: "Price Low to High",
    pricehigh: "Price High to Low",
};

export

function getSearchOptionsFromURL() {
    "use strict";
    var options = urlQuery();

    options.search = decodeURIComponent(options.search || "");
    options.towns = [].concat(options.town || []);
    options.tags = [].concat(options.tag || []);
    options.category = options.category || "all";

    return options;
}

export

function buildSearchParams(options) {
    "use strict";
    var query = {};

    if (options.towns) {
        query.town = options.towns;
    }

    if (options.category && options.category != "all") {
        query.category = options.category;
    }

    if (options.search) {
        query.search = encodeURIComponent(options.search);
    }

    if (options.tags) {
        query.tag = options.tags;
    }

    if (options.since) {
        if (typeof options.since == "object") {
            query.since = options.since.toJSON();
        } else {
            query.since = options.since;
        }
    }

    if (options.from) {
        query.from = options.from;
    }

    if (options.latitude) {
        query.latitude = options.latitude;
    }

    if (options.longitude) {
        query.longitude = options.longitude;
    }

    if (options.milesDistant) {
        query.milesDistant = options.milesDistant;
    }

    if (options.sort) {
        query.sort = options.sort;
    }

    if (options.minPrice) {
        query.minPrice = options.minPrice;
    }

    if (options.maxPrice) {
        query.maxPrice = options.maxPrice;
    }

	if(options.showModerated) {
		query.showModerated = options.showModerated;
	}


    query.limit = options.limit || resultLimit;

    return $.param(query, true);
}

export

function posts(options) {
    "use strict";

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/posts/?" + buildSearchParams(options),
    });

}

export

function search(options) {
    "use strict";

    return csrf.ajax({
        type: "GET",
        url: "/api/v1/search?" + buildSearchParams(options),
    });
}
