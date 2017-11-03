// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

import {
    origin,
    fnv32,
}
from "./ts/util";

import {
    bookmarkletElementID,
    autoParse,
}
from "./ts/bookmarklet";

(function() {
    "use strict";

    // unique id to prevent conflics
    var id = fnv32(window.location.toString());

    var parsed = autoParse();

    if (parsed) {
        share();
        return;
    }

    parsed = {
        title: "",
        content: "",
        images: [],
        imageKeys: [],
    };


    var selected = [];

    parsed.title = document.title;

    // else client side pick
    cleanUp(); // make sure any previous bookmarklet runs are cleaned up
    addCSS();
    addTSBar();

    var overlay = addOverlay();

    window.addEventListener("resize", reset, false);

    function share() {
        var u = tsDomain() + "/share/?url=" + encodeURIComponent(location);

        if (parsed.title) {
            u += "&title=" + encodeURIComponent(parsed.title);
        }

        if (parsed.content) {
            u += "&content=" + encodeURIComponent(parsed.content);
        }

        if (parsed.images && parsed.images.length) {
            for (var i = 0; i < parsed.images.length; i++) {
                u += "&image=" + encodeURIComponent(parsed.images[i]);
            }
        }

        if (parsed.imageKeys && parsed.imageKeys.length) {
            for (var j = 0; j < parsed.imageKeys.length; j++) {
                u += "&imagekeys=" + encodeURIComponent(parsed.imagekeys[j]);
            }
        }

        window.location = u;
    }

    function tsDomain() {
        var s = document.getElementById(bookmarkletElementID).getAttribute("src").split("/");

        return s[0] + "//" + s[2];
    }

    function addCSS() {
        var css = "#" + getName("overlay") + "{z-index: 2147483646; position: absolute; top: 0; left: 0;}" +
            "html { margin-top: 52px !important;}" +
            "#" + getName("tsBar") + "{z-index: 2147483647; position: fixed; top: 0; left:0;right:0;height: 52px; " +
            "margin-top: -52px;transition: all 1s ease; background-color: #6B9F3B; color: white; text-align: center;}" +
            "#" + getName("tsBar") + " * {font-size: 18px; padding: 0; margin: 14px 0 0 0; font-weight: bold; " +
            "font-family: 'Helvetica Neue', Arial, Helvetica, sans-serif; }" +
            "#" + getName("tsBar") + " a {text-decoration: none;}" +
            "#" + getName("tsBar") + " a:hover {text-decoration: underline;}" +
            "#" + getName("tsIcon") + "{float: left; stroke: white; fill: white; height: 42px; width: 42px; margin: 5px 0 0 5px;}" +
            "." + getName("can-select") + "{ opacity: 0; transition: all .3s ease; fill: white;stroke: white; cursor: copy;}" +
            "." + getName("can-select") + ":hover{ opacity: 0.5;}" +
            "." + getName("selected") + "{opacity: 0;fill: black;stroke: black;cursor: default;}" +
            "." + getName("selected") + ":hover{ opacity: 0;}" +
            "." + getName("hide") + "{display: none;}" +
            "." + getName("info") + "{ display: inline-block;}" +
            "#" + getName("next") + "{ float: right; width: 100px; font-size: 22px; color: white; margin: 10px 0 0 0;}" +
            "#" + getName("cancel") + "{ float: left; width: 100px; font-size: 12px; color: white; margin: 17px 0 0 0;}";
        var head = document.getElementsByTagName("head").item(0);
        var s = document.createElement("style");
        markForCleanup(s);
        s.type = "text/css";
        s.appendChild(document.createTextNode(css));
        head.appendChild(s);
    }

    function addOverlay(nodes, selected) {
        var size = {
            height: document.documentElement.scrollHeight,
            width: document.documentElement.scrollWidth,
        };

        var ol = document.createElementNS("http://www.w3.org/2000/svg", "svg");
        ol.setAttribute("id", getName("overlay"));
        ol.setAttributeNS(null, "width", size.width);
        ol.setAttributeNS(null, "height", size.height);
        ol.setAttributeNS(null, "viewBox", "0 0 " + size.width + " " + size.height);
        markForCleanup(ol);

        var defs = document.createElementNS("http://www.w3.org/2000/svg", "defs");
        var mask = document.createElementNS("http://www.w3.org/2000/svg", "mask");
        mask.setAttributeNS(null, "id", "selected");

        var maskR = document.createElementNS("http://www.w3.org/2000/svg", "rect");
        maskR.setAttributeNS(null, "width", size.width);
        maskR.setAttributeNS(null, "height", size.height);
        maskR.setAttributeNS(null, "x", "0");
        maskR.setAttributeNS(null, "y", "0");
        maskR.setAttributeNS(null, "fill", "white");

        mask.appendChild(maskR);

        defs.appendChild(mask);

        ol.appendChild(defs);

        var r = document.createElementNS("http://www.w3.org/2000/svg", "rect");
        r.setAttributeNS(null, "width", size.width);
        r.setAttributeNS(null, "height", size.height);
        r.setAttributeNS(null, "x", "0");
        r.setAttributeNS(null, "y", "0");
        r.setAttributeNS(null, "fill", "black");
        r.setAttributeNS(null, "opacity", "0.25");
        r.setAttributeNS(null, "mask", "url(#selected)");
        ol.appendChild(r);

        document.body.appendChild(ol);

        ol = {
            el: ol,
            nodes: nodes || getNodes(),
            selected: [],
            toggleSelect: function(el) {
                var dataIndex = el.getAttribute("data-index");
                var index = this.selected.indexOf(dataIndex);

                if (index < 0) {
                    // add selected class
                    this.select(el);
                    return;

                }
                el.classList.remove(getName("selected"));
                ol.el.getElementById("selected").removeChild(ol.el.getElementById(this.maskID(dataIndex)));
                this.selected.splice(index, 1);
            },
            maskID: function(index) {
                return "maskSelect-" + index;
            },
            select: function(el) {
                el.classList.add(getName("selected"));
                var dataIndex = el.getAttribute("data-index");
                var mask = getRect(this.nodes[dataIndex]);
                mask.setAttributeNS(null, "id", this.maskID(dataIndex));
                mask.setAttributeNS(null, "stroke", "black");
                mask.setAttributeNS(null, "fill", "black");

                this.el.getElementById("selected").appendChild(mask);

                this.selected.push(dataIndex);
            },
            addNode: function(n) {
                this.nodes.push(n);
                var item = getRect(n);

                item.setAttributeNS(null, "class", getName("can-select"));
                item.setAttributeNS(null, "data-index", (this.nodes.length - 1));

                item.onclick = toggleSelect;
                this.el.appendChild(item);


            },
        };

        for (var i = 0; i < ol.nodes.length; i++) {
            var item = getRect(ol.nodes[i]);

            item.setAttributeNS(null, "class", getName("can-select"));
            item.setAttributeNS(null, "data-index", i);

            item.onclick = toggleSelect;
            ol.el.appendChild(item);

            if (selected) {
                var selIndex = selected.indexOf(i.toString());
                if (selIndex >= 0) {
                    ol.select(item);
                }
            }

        }

        return ol;
    }

    function toggleSelect(event) {
        event.preventDefault();
        event.stopPropagation();
        overlay.toggleSelect(event.target);
    }


    function getName(name) {
        return "ts-" + id + "-" + name;
    }

    // mark the given element for clean by adding a unique classname to it
    function markForCleanup(element) {
        element.classList.add(getName("bookmarklet-item"));
    }

    function cleanUp() {
        var els = document.getElementsByClassName(getName("bookmarklet-item"));
        var len = els.length;

        for (var i = 0; i < len; i++) {
            var parent = els[0].parentNode;
            parent.removeChild(els[0]);
        }

        window.removeEventListener("resize", reset, false);
    }


    function getNodes() {
        var text = [];
        var images = [];
        var treeWalker = document.createTreeWalker(document.body);
        //TODO: Fix performance
        // only do elements on the current scroll page?
        var max = 1000;

        while (treeWalker.nextNode()) {
            var n = treeWalker.currentNode;
            max--;
            if (max <= 0) {
                break;
            }
            if (n.nodeType == Node.ELEMENT_NODE) {
                if (n.nodeName.toUpperCase() == "IMG") {
                    images.push(n);
                    continue;
                }

                if (n.nodeName.toUpperCase() == "UL" || n.nodeName.toUpperCase() == "OL") {
                    text.push(n);
                    // grabbed the whole tree from here, so we'll just skip to the next sibling
                    treeWalker.nextSibling();
                    treeWalker.previousNode(); // so sibling doesn't get skipped on loop
                    continue;
                }
            }

            if (n.nodeType == Node.TEXT_NODE) {
                if (n.textContent.trim()) {
                    text.push(n.parentNode);
                }
                continue;
            }


        }

        return text.concat(images);
    }

    function getRect(el) {

        var pos = el.getBoundingClientRect();

        var s = document.createElementNS("http://www.w3.org/2000/svg", "rect");
        s.setAttributeNS(null, "width", pos.right - pos.left);
        s.setAttributeNS(null, "height", pos.bottom - pos.top);
        s.setAttributeNS(null, "x", pos.left + window.scrollX);
        s.setAttributeNS(null, "y", pos.top + window.scrollY);
        s.setAttributeNS(null, "ry", "4");
        s.setAttributeNS(null, "rx", "4");
        s.setAttributeNS(null, "stroke-width", "10");

        return s;
    }

    function addTSBar() {
        var bar = document.createElement("div");
        markForCleanup(bar);
        bar.setAttribute("id", getName("tsBar"));

        // ts icon and link
        var tsLink = document.createElement("a");
        tsLink.setAttribute("href", tsDomain());
        tsLink.setAttribute("title", "go to townsourced.com");
        addIcon(tsLink);
        bar.appendChild(tsLink);

        // Next link
        var next = document.createElement("a");
        next.setAttribute("id", getName("next"));
        next.appendChild(document.createTextNode("Next"));
        next.setAttribute("href", "#");
        next.onclick = function(event) {
            event.preventDefault();
            parseSelection();
            share();
        };
        bar.appendChild(next);

        // cancel link
        var cancel = document.createElement("a");
        cancel.setAttribute("id", getName("cancel"));
        cancel.appendChild(document.createTextNode("Cancel"));
        cancel.setAttribute("href", "#");
        cancel.onclick = function(event) {
            event.preventDefault();
            cleanUp();
            // remove script element
            document.documentElement.removeChild(document.getElementById(bookmarkletElementID));
        };
        bar.appendChild(cancel);


        // info text
        var info = document.createElement("span");
        info.classList.add(getName("info"));
        info.appendChild(document.createTextNode("Select any text or image you want included in your post"));
        bar.appendChild(info);


        document.body.appendChild(bar);
        // make sure initial margin is set
        var nothing = window.getComputedStyle(bar)["margin-top"];

        bar.style["margin-top"] = "0";
    }

    var resizeDelay;

    function reset() {
        if (resizeDelay) {
            window.clearTimeout(resizeDelay);
        }
        resizeDelay = window.setTimeout(function() {
            document.body.removeChild(overlay.el);
            overlay = addOverlay(overlay.nodes, overlay.selected);
        }, 500);
    }

    function parseSelection() {
        for (var i = 0; i < overlay.nodes.length; i++) {
            var n = overlay.nodes[i];

            if (overlay.selected.indexOf(i.toString()) > -1) {
                if (n.nodeType == Node.ELEMENT_NODE && n.nodeName.toUpperCase() == "IMG") {
                    parsed.images.push(n.getAttribute("src"));
                } else {
                    var val = n.outerHTML.trim();
                    if (val) {
                        parsed.content += n.outerHTML;
                    }
                }
            }
        }
    }

    function addIcon(parent) {
        var icon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
        icon.setAttribute("id", getName("tsIcon"));
        icon.setAttributeNS(null, "width", "120");
        icon.setAttributeNS(null, "height", "120");
        icon.setAttributeNS(null, "viewBox", "0 0 120 120");

        var path1 = document.createElementNS("http://www.w3.org/2000/svg", "path");
        path1.setAttributeNS(null, "d", "m1.4187 1.4851 3.7098 115.16 113.46 1.85-0.93-117.01h-116.25zm108.84 10.071l0.29391 97.37-97.077 0.14799 1.0297-90.621 95.753-6.8975z");
        path1.setAttributeNS(null, "fill-rule", "evenodd");
        path1.setAttributeNS(null, "stroke-width", "1px");
        icon.appendChild(path1);

        var path2 = document.createElementNS("http://www.w3.org/2000/svg", "path");
        path2.setAttributeNS(null, "d", "m46.653 83.447q2.2696 0 4.287-0.50435 2.0678-0.50435 4.1861-1.2609v11.449q-2.1687 1.1096-5.3965 1.8157-3.1774 0.75652-6.96 0.75652-3.6817 0-6.8591-0.85739t-5.4974-2.9757q-2.32-2.1687-3.6817-5.7496-1.3113-3.6313-1.3113-9.0783v-27.184h-7.3635v-6.5061l8.4731-5.1444 4.4383-11.903h9.8348v12.003h13.718v11.55h-13.718v27.184q0 3.2783 1.6139 4.8417 1.6139 1.5635 4.2365 1.5635z");
        icon.appendChild(path2);

        var path3 = document.createElementNS("http://www.w3.org/2000/svg", "path");
        path3.setAttributeNS(null, "d", "m104.1 77.949q0 4.4383-1.6139 7.767t-4.5896 5.5478-7.2122 3.3287-9.4818 1.1096q-2.7739 0-5.1444-0.20174-2.3704-0.1513-4.4887-0.55478t-4.0852-1.0087q-1.967-0.60522-3.9844-1.513v-12.71q2.1183 1.0591 4.4383 1.9165 2.3704 0.85739 4.6904 1.513 2.32 0.60522 4.4887 0.95826 2.2191 0.35304 4.0852 0.35304 2.0678 0 3.5304-0.35304 1.4626-0.40348 2.3704-1.0591 0.95826-0.70609 1.3617-1.6139 0.45391-0.95826 0.45391-2.0174t-0.35304-1.8661q-0.30261-0.85739-1.4626-1.7652-1.16-0.95826-3.4296-2.1183-2.2191-1.2104-6.0017-2.9252-3.6817-1.6644-6.4052-3.2783-2.673-1.6644-4.4383-3.6817-1.7148-2.0174-2.5722-4.5896-0.85739-2.6226-0.85739-6.2035 0-3.9339 1.513-6.8591 1.513-2.9757 4.287-4.9426 2.7739-1.967 6.6574-2.9252 3.9339-1.0087 8.7252-1.0087 5.0435 0 9.5826 1.16t9.3304 3.48l-4.64 10.894q-3.833-1.8157-7.3131-2.9757-3.48-1.16-6.96-1.16-3.127 0-4.5391 1.1096-1.3617 1.1096-1.3617 3.0261 0 1.0087 0.35304 1.8157 0.35304 0.75652 1.4122 1.6139 1.0591 0.80696 2.9757 1.8157 1.9165 0.95826 4.9931 2.3704 3.5809 1.5635 6.4557 3.127 2.8748 1.513 4.9426 3.48t3.1774 4.5896 1.1096 6.3548z");
        icon.appendChild(path3);
        parent.appendChild(icon);
    }


}());
