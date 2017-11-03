// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

var defaults;

defaults = {
    duration: 250,
    easing: 'ease-out',
    fade: true,
    from: 0.3,
    to: 1
};

export
function scale(t, params) {
    'use strict';
    params = t.processParams(params, defaults);

    var scaleTo = 'scale(' + params.to + ')',
        scaleFrom = 'scale(' + params.from + ')',
        targetOpacity, anim = {};

    if (t.isIntro) {
        t.setStyle('transform', scaleFrom);

        if (t.fade !== false) {
            targetOpacity = t.getStyle('opacity');
            t.setStyle('opacity', 0);
        }
    }

    // set defaults
    anim.opacity = t.isIntro ? targetOpacity : 0;

    if (t.fade !== false) anim.transform = t.isIntro ? scaleTo : scaleFrom;

    t.animateStyle(anim, params).then(t.complete);
    // as of 0.4.0 transitions return a promise and transition authors should do t.anymateStyle(params).then(action)
}
