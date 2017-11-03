// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
var page;
export

function err(response, auto) {
    "use strict";

    var error = {
        message: null,
    };

    if (!response) {
        return error;
    }

    if (typeof response === "string") {
        error.message = response;

        return error;
    } else {
		error.message = "An error occurred";

        if (!response.responseJSON) {
            return error;
        }

        var statusCode = response.status;
        response = response.responseJSON;
        error.message = response.message || error.message;
        error.statusCode = statusCode;


        if (response.status == "error") {
            if (page) {
                page.fire("error");
            }
            return error;
        } else {
            //failures
            if (statusCode == 404) {
                error.notFound = true;
                if (auto) {
                    if (page) {
                        page.fire("four04", error);
                    }
                    return error;
                }

            }
            if (statusCode == 409) {
                error.conflict = true;
                if (page) {
                    page.fire("versionConflict", error);
                }
                return error;
            }
            if (statusCode == 429) {
                error.tooManyRequests = true;
                if (page) {
                    page.fire("tooManyRequests", error);
                }
                return error;
            }
            if (error.message == "Invalid CSRFToken") {
                error.csrf = true;
                if (page) {
                    page.fire("invalidCsrf", error);
                }
                return error;
            }

            return error;
        }
    }
}

export

function setPage(ractivePage) {
    "use strict";
    page = ractivePage;
}
