// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

var DEFAULTS = {
    duration: 500,
    easing: 'easeInOut',
    minDuration: 200,
    maxDuration: 500,
};

var PROPS = [
    "-webkit-transform-style",
    "transform-style",
    "-webkit-transform",
    "transform",
];

var COLLAPSED = {
    "-webkit-transform-style": "preserve-3d",
    "transform-style": "preserve-3d",
    "-webkit-transform": "translateZ(400px) translateY(300px) rotateX(-90deg)",
    "transform": "translateZ(400px) translateY(300px) rotateX(-90deg)",
};

export
default

function slide(t, params) {
    "use strict";
    var targetStyle;

    if (!t.node.parentNode.style["-webkit-perspective"] && !t.node.parentNode.style.perspective) {
        t.node.parentNode.style["-webkit-perspective"] = "1300px";
        t.node.parentNode.style.perspective = "1300px";
    }

    params = t.processParams(params, DEFAULTS);

    if (params.minDuration) {
        var maxDuration = params.maxDuration || params.duration;
        params.duration = (Math.random() * (maxDuration - params.minDuration) + params.minDuration);
    }

    if (t.isIntro) {
        targetStyle = t.getStyle(PROPS);
        t.setStyle(COLLAPSED);
    } else {
        // make style explicit, so we're not transitioning to 'auto'
        t.setStyle(t.getStyle(PROPS));
        targetStyle = COLLAPSED;
    }

    t.animateStyle(targetStyle, params).then(t.complete);
}
