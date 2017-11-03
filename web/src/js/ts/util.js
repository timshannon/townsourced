// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  strict: true, esnext: true*/

export

function origin() {
    "use strict";
    if (!window.location.origin) {
        //for IE
        window.location.origin = window.location.protocol + "//" + window.location.hostname +
            (window.location.port ? ':' + window.location.port : '');
    }
    return location.origin;
}


export

function overflowToTitle() {
    "use strict";
    $(document).on('mouseenter', '.overflow-title', function() {
        var e = $(this);

        if (this.offsetWidth < this.scrollWidth) {
            if (!e.attr("title") || e.attr("title") != e.text()) {
                e.attr("title", e.text());
            }
        } else {
            e.removeAttr("title");
        }
    });
}

export

function urlify(value) {
    "use strict";
    if (!value) {
        return value;
    }

    //FIXME fix space handling on newtown entry
    return value.trim().replace(/[^a-zA-Z0-9\-]/g, "-").toLowerCase();
}

export

function htmlPayload(id) {
    "use strict";
    id = id || "payload";
    var el = document.getElementById(id);
    if (!el) {
        return null;
    }

    if (!el.innerHTML) {
        return null;
    }

    if (el.innerHTML.trim() === "") {
        return null;
    }
    return JSON.parse(el.innerHTML);
}

export

function escapeRegExp(string) {
    "use strict";
    return string.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}


export

function urlQuery() {
    "use strict";
    var query_string = {};

    var query = window.location.search.substring(1);
    var vars = query.split("&");
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split("=");
        // If first entry with this name
        if (typeof query_string[pair[0]] === "undefined") {
            query_string[pair[0]] = decodeURIComponent(pair[1]);
            // If second entry with this name
        } else if (typeof query_string[pair[0]] === "string") {
            var arr = [query_string[pair[0]], decodeURIComponent(pair[1])];
            query_string[pair[0]] = arr;
            // If third or later entry with this name
        } else {
            query_string[pair[0]].push(decodeURIComponent(pair[1]));
        }
    }
    return query_string;
}

export

function urlJoin() {
    "use strict";
    var j = [].slice.call(arguments, 0).join('/');

    return j.replace(/[\/]+/g, '/')
        .replace(/\/\?/g, '?')
        .replace(/\/\#/g, '#')
        .replace(/\:\//g, '://');
}

export

function unique(arr, optKey) {
    "use strict";
    var a = [];
    var i = 0;

    if (optKey && arr[0] && typeof arr[0] == "object") {
        var s = [];
        for (i = 0; i < arr.length; i++) {
            if (arr[i] === null) {
                continue;
            }
            if (s.indexOf(arr[i][optKey]) === -1 && arr[i] !== '') {
                a.push(arr[i]);
                s.push(arr[i][optKey]);
            }
        }
        return a;
    }

    for (i = 0; i < arr.length; i++) {
        if (arr[i] === null) {
            continue;
        }

        if (a.indexOf(arr[i]) === -1 && arr[i] !== '') {
            a.push(arr[i]);
        }
    }
    return a;
}

export

function findIndex(arr, src, optKey) {
    "use strict";
    if (!optKey || (arr[0] && typeof arr[0] != "object")) {
        return arr.indexOf(src);
    }

    var match = src;

    if (typeof src == "object" && optKey) {
        match = src[optKey];
    }

    if (!match) {
        return -1;
    }


    for (var i = 0; i < arr.length; i++) {
        if (arr[i][optKey] && arr[i][optKey] == match) {
            return i;
        }
    }
    return -1;

}


export

function pluck(arr, key) {
    "use strict";
    var result = [];

    for (var i = 0; i < arr.length; i++) {
        result.push(arr[i][key]);
    }

    return result;
}


export

function find(arr, src, optKey) {
    "use strict";
    return arr[findIndex(arr, src, optKey)];
}

export

function loadAndStore(ractive, keys, keypath, loadFunc, optAltSource) {
    "use strict";

    if (!keys) {
        return;
    }

    if (typeof keys == "string") {
        keys = [keys];
    }

    if (!ractive) {
        throw "Ractive instance is required";
    }


    var getItem = function(key, keypath) {
        loadFunc(key)
            .done(function(result) {
                ractive.set(keypath + "." + key, result.data);
            });
    };


    var altSource;

    if (!optAltSource) {
        altSource = {};
    } else {
        if (Array.isArray(optAltSource)) {
            altSource = {};
            $.each(optAltSource, function(key, val) {
                altSource[val.key] = val;
            });
        } else {
            altSource = optAltSource;
        }
    }

    for (var i = 0; i < keys.length; i++) {
        if (ractive.get(keypath + "." + keys[i])) {
            continue;
        }


        var item = altSource[keys[i]];

        if (!item) {
            //didn't find locally, load it from api
            item = {
                key: keys[i],
                name: keys[i],
                loading: true,
            };

            getItem(keys[i], keypath);
        }

        ractive.set(keypath + "." + keys[i], item);
    }
}

export

function since(strDate) {
    "use strict";
    var date = new Date(strDate);

    if (!date) {
        return "";
    }



    var seconds = Math.floor((new Date() - date) / 1000);

    var interval = Math.floor(seconds / 31536000);

    if (interval > 1) {
        return interval + " years";
    }
    interval = Math.floor(seconds / 2592000);
    if (interval > 1) {
        return interval + " months";
    }
    interval = Math.floor(seconds / 86400);
    if (interval > 1) {
        return interval + " days";
    }
    interval = Math.floor(seconds / 3600);
    if (interval > 1) {
        return interval + " hours";
    }
    interval = Math.floor(seconds / 60);
    if (interval > 1) {
        return interval + " minutes";
    }
    return Math.floor(seconds) + " seconds";
}

export

function formatDate(strDate) {
    "use strict";
    var date = new Date(strDate);

    if (!date) {
        return "";
    }

    return date.toLocaleDateString() + " at " + date.toLocaleTimeString();
}


var scrollNodes = [];

export

function scrollToFixed(node, optTrigger, optClass, optFunc) {
    "use strict";

    scrollNodes.push({
        node: $(node),
        trigger: optTrigger || node,
        class: optClass || "fixed",
        func: optFunc || null,
    });

    if (scrollNodes.length == 1) {
        $(window).scroll(function() {
            for (var i = 0; i < scrollNodes.length; i++) {
                if (!scrollNodes[i].top) {
                    scrollNodes[i].top = $(scrollNodes[i].trigger).offset().top;
                }
                var n = scrollNodes[i];

                if ($(window).scrollTop() > n.top) {
                    if (n.func) {
                        n.func(true);
                    } else {
                        if (!n.node.hasClass(n.class)) {
                            n.node.addClass(n.class);
                        }
                    }
                } else {
                    if (n.func) {
                        n.func(false);
                    } else {
                        if (n.node.hasClass(n.class)) {
                            n.node.removeClass(n.class);
                        }
                    }
                }
            }
        });
    }
}


export

// fetchFunc(lastItem, dataLength)
function addPager(r) {
    "use strict";
    r.set("onPage", function(keypath, index) {
        var p = this.get(keypath);

        var start = (p.pageSize * p.page) - p.pageSize;
        var end = (p.pageSize * p.page) - 1;

        return (index >= start && index <= end);
    });
    r.set("lastPage", function(keypath, optPage) {
        var p = this.get(keypath);
        var page = optPage || p.page;
        if (!p.data) {
            return true;
        }

        return lastPage(page, p.pageSize, p.data.length);
    });

    r.nextPage = function(keypath) {
        this.add(keypath + ".page");
        var p = this.get(keypath);

        // if next to last page load more
        if (lastPage(p.page + 1, p.pageSize, p.data.length) && p.fetch) {
            p.fetch(p.data[p.data.length - 1], p.data.length);
        }
    };
    r.prevPage = function(keypath) {
        this.add(keypath + ".page", -1);
    };

    function lastPage(page, pageSize, dataLength) {
        return (page * pageSize) >= dataLength;
    }
}

export

function isEmail(str) {
    "use strict";
    return /.+@.+\..+/i.test(str);
}

export

function isURL(str) {
    "use strict";
    return /(http|https):\/\/[\w-]+(\.[\w-]+)+([\w.,@?^=%&amp;:/~+#-]*[\w@?^=%&amp;/~+#-])?/i.test(str);
}


export

function endsWith(s, search, position) {
    "use strict";

    var subject = s.toString();
    if (typeof position !== 'number' || !isFinite(position) || Math.floor(position) !== position || position > subject.length) {
        position = subject.length;
    }
    position -= search.length;
    var lastIndex = subject.indexOf(search, position);

    return lastIndex !== -1 && lastIndex === position;
}

//https://gist.github.com/vaiorabbit/5657561
export

function fnv32(str, asString, seed) {
    "use strict";
    /*jshint bitwise:false */
    var i, l,
        hval = (seed === undefined) ? 0x811c9dc5 : seed;

    for (i = 0, l = str.length; i < l; i++) {
        hval ^= str.charCodeAt(i);
        hval += (hval << 1) + (hval << 4) + (hval << 7) + (hval << 8) + (hval << 24);
    }
    if (asString) {
        // Convert to 8 digit hex string
        return ("0000000" + (hval >>> 0).toString(16)).substr(-8);
    }
    return hval >>> 0;
}
