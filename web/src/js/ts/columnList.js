// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

export
default

function(parent, itemWidth, gutter, optOutsideGutter, sidebar) {
    "use strict";
    var lHeight = 0;
    var lWidth = 0;
    var itemCount = 0;

    this.sidebar = sidebar;
    this.columns = [];
    this.parent = parent;
    this.itemWidth = itemWidth;
    this.gutter = gutter;
    this.outsideGutter = optOutsideGutter || this.gutter;
    this.listHeight = function() {
        return lHeight;
    };
    this.fitsInWindow = function() {
        return this.columnCount() == this.columns.length;
    };
    this.listWidth = function() {
        return lWidth;
    };
    this.reset = function(preserveParentHeight) {
        // determine the number of columns that can fit on screen, and set their
        // starting heights at 0;
        this.columns = [];
        lHeight = 0;
        lWidth = 0;
        itemCount = 0;

        if (!preserveParentHeight) {
            this.parent.style.height = lHeight + "px";
            this.parent.style.width = lWidth + "px";
        }
        var n = this.columnCount();

        for (var i = 0; i < n; i++) {
            this.columns.push(0);
        }
    };
    this.columnCount = function() {
        var w = window.innerWidth - (this.outsideGutter * 2);

        // gutter is only in between columns, and not on the outer left or right of the
        // column group
        // can't be less than 1 column
        return Math.max(1, Math.floor(w / (this.itemWidth + this.gutter)));
    };
    this.placeAtNextPosition = function(node) {
        var pos = this.nextPosition(node.getBoundingClientRect().height);

        node.style.top = pos.top + "px";
        node.style.left = pos.left + "px";
    };
    this.nextPosition = function(curItemHeight) {
        var pos = {
            top: -1,
            left: -1,
            col: -1,
        };

        for (var i = 0; i < this.columns.length; i++) {
            /*if (this.sidebar && i === 0 && this.columns.length > 1) {*/
            /*continue;*/
            /*}*/
            if (this.columns[i] < pos.top || pos.top == -1) {
                pos.top = this.columns[i];
                pos.col = i;
            }
        }



        pos.top += this.gutter;
        pos.left = pos.col * (this.itemWidth + this.gutter);
        this.columns[pos.col] += (curItemHeight + this.gutter);


        if (this.columns[pos.col] > lHeight) {
            lHeight = this.columns[pos.col];
            this.parent.style.height = lHeight + "px";
        }

        itemCount++;

        if (itemCount <= this.columns.length) {
            //min between column count or item count, so that list can
            // be centered properly on item counts < column count

            lWidth = ((this.itemWidth + this.gutter) *
                Math.min(itemCount, this.columns.length)) - this.gutter;
            this.parent.style.width = lWidth + "px";
        }

        if (this.sidebar && pos.col === 0 && this.columns.length > 1) {
            return this.nextPosition(curItemHeight);
        }

        return pos;
    };
    this.itemCount = function() {
        return itemCount;
    };

    this.reset();


}
