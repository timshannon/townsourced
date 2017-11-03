// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */


var props, collapsed, defaults;

defaults = {
    duration: 300,
    easing: 'easeInOut'
};

props = [
    'width',
    'borderLeftWidth',
    'borderRightWidth',
    'paddingLeft',
    'paddingRight',
    'marginLeft',
    'marginRight'
];

collapsed = {
    width: 0,
    borderLeftWidth: 0,
    borderRightWidth: 0,
    paddingLeft: 0,
    paddingRight: 0,
    marginLeft: 0,
    marginRight: 0
};

export
function slide(t, params) {
    "use strict";
    var targetStyle;

    params = t.processParams(params, defaults);

    if (t.isIntro) {
        t.setStyle('height', t.getStyle('height'));
        targetStyle = t.getStyle(props);
        t.setStyle(collapsed);
    } else {
        //make style explicit, so we're not transitioning to 'auto'
        t.setStyle('height', t.getStyle('height'));
        t.setStyle(t.getStyle(props));
        targetStyle = collapsed;
    }
    t.setStyle({
        overflow: 'hidden'
    });

    t.animateStyle(targetStyle, params).then(t.complete);
}
