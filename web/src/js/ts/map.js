// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  strict: true, esnext: true*/

import {
    get as getStorage
}
from "./storage";

export

function newMap(node, optShowType) {
    "use strict";
    var latlng = new google.maps.LatLng(getStorage("lastLat") || 39.0695144, getStorage("lastLng") || -96.4315569);
    var mapOptions = {
        zoom: 5,
        minZoom: 2,
        center: latlng,
        disableDefaultUI: true,
        mapTypeId: google.maps.MapTypeId.ROADMAP,
        mapTypeControl: optShowType,
        mapTypeControlOptions: {
            position: google.maps.ControlPosition.RIGHT_TOP,
        },
        zoomControl: true,
        zoomControlOptions: {
            position: google.maps.ControlPosition.LEFT_CENTER,
        },
        scaleControl: true,
    };

    return new google.maps.Map(node, mapOptions);
}

export
var secondaryMarker = {
    path: "m4.5746-18.285q0-1.893-1.339-3.232t-3.232-1.339-3.232 1.339-1.339 3.232 1.339 3.232 3.232 1.339 3.232-1.339 1.339-3.232zm4.572 0q0 1.946-0.589 3.196l-6.5 13.821q-0.286 0.589-0.848 0.929-0.56196 0.34-1.205 0.339-0.64296-0.0010029-1.205-0.339-0.562-0.338-0.83-0.929l-6.518-13.821q-0.589-1.25-0.589-3.196 0-3.786 2.679-6.464t6.464-2.679 6.464 2.679 2.679 6.464z",
    fillOpacity: 0.8,
    scale: 1,
    fillColor: "#777",
    strokeWeight: 0,
};


// returns the distance between two latlgns
export

function haversine(lat1, lng1, lat2, lng2) {
    "use strict";

    var rad = function(x) {
        return x * Math.PI / 180;
    };

    var dLat = rad(lat2 - lat1);
    var dLon = rad(lng2 - lng1);

    var a = Math.sin(dLat / 2) * Math.sin(dLat / 2) + Math.sin(dLon / 2) * Math.sin(dLon / 2) * Math.cos(rad(lat1)) *
        Math.cos(rad(lat2));
    var c = 2 * Math.asin(Math.sqrt(a));

    return 6372.8 * c;
}
