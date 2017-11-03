// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true */

var csrfToken;
export
default {
    set: function(xhr) {
        xhr.setRequestHeader("X-CSRFToken", csrfToken);
    },
    get: function(xhr) {
		if(typeof xhr == "string") {
			csrfToken = xhr;
			return;
		}
        var token = xhr.getResponseHeader("X-CSRFToken");
        if (token) {
            csrfToken = token;
        }
    },
    ajax: function(type, url, options) {
        if (typeof type !== "string") {
            options = type;
        }
        if (!options) {
            options = {};
        }
        if (!options.type) {
            options.type = type;
        }
        if (!options.url) {
            options.url = url;
        }
        if (!options.dataType) {
            options.dataType = "json";
        }

        var csrf = this;
        if (options.type != "GET") {
            if (options.beforeSend) {
                var origBefore = options.beforeSend;
                options.beforeSend = function(xhr) {
                    csrf.set(xhr);
                    origBefore(xhr);
                };
            } else {
                options.beforeSend = this.set;
            }
        } else {
            if (options.complete) {
                var origComplete = options.complete;
                options.complete = function(xhr) {
                    csrf.get(xhr);
                    origComplete(xhr);
                };
            } else {
                options.complete = this.get;
            }
        }
        return $.ajax(options);
    },
};
