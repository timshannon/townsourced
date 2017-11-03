// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true*/

import Ractive from "ractive";
import {
    newMap,
    secondaryMarker,
}
from "../../ts/map";
import {
    set as setStorage
}
from "../../ts/storage";


export
default {
    decorators: {
        map: function(node) {
            var r = this;

            var map = newMap(node, !r.get("showRange"));
            var searchBox = new google.maps.places.SearchBox(r.nodes.searchBox);

            // Bias the SearchBox results towards current map's viewport.
            map.addListener("bounds_changed", function() {
                var searchBox = r.get("searchBox");
                var map = r.get("map");
                if (searchBox && map) {
                    searchBox.setBounds(map.getBounds());
                }
            });

            searchBox.addListener("places_changed", function() {
                var searchBox = r.get("searchBox");
                var map = r.get("map");
                if (!searchBox || !map) {
                    return;
                }

                var places = searchBox.getPlaces();

                if (places.length === 0) {
                    return;
                }

                r.fire("clearSecondaryMarkers");

                var bounds = new google.maps.LatLngBounds();
                var first;

                places.forEach(function(place) {
                    if (place.geometry.viewport) {
                        // Only geocodes have viewport.
                        bounds.union(place.geometry.viewport);
                    } else {
                        bounds.extend(place.geometry.location);
                    }
                    if (!first) {
                        first = places.pop();
                        r.fire("addSecondaryMarker", place.geometry.location, place.name, true);
                    } else {
                        r.fire("addSecondaryMarker", place.geometry.location, place.name);
                    }
                });

                map.fitBounds(bounds);

                r.fire("setLocation", first.geometry.location.lat(), first.geometry.location.lng());
            });


            r.set("map", map);
            r.set("searchBox", searchBox);
            return {
                teardown: function() {
                    r.set("map", undefined);
                    r.set("marker", undefined);
                    r.set("searchBox", undefined);
                },
            };
        },
    },
    isolated: true,
    data: function() {
        return {
            latitude: 0,
            longitude: 0,
            searchString: "",
            range: 3,
            ranges: [
                1, 3, 5, 10, 20, 30, 40, 50, 75, 100
            ],
            markers: [],
        };
    },
    onrender: function() {
        var r = this;

        r.on({
            "search": function(event) {
                event.original.preventDefault();
                if (!event.context.searchString) {
                    return;
                }

                r.set("search", event.context.searchString);
                r.set("latitude", null);
                r.set("longitude", null);
                r.set("error", null);

                var searchBox = r.get("searchBox");
                if (!searchBox) {
                    return;
                }

                google.maps.event.trigger(r.nodes.searchBox, 'focus');
                google.maps.event.trigger(r.nodes.searchBox, 'keydown', {
                    keyCode: 13
                });


            },
            "reset": function() {
                r.set("search", null);
                r.set("searchString", null);
                r.set("latitude", null);
                r.set("longitude", null);

                r.set("error", null);
                var map = r.get("map");
                if (map) {
                    google.maps.event.trigger(map, 'resize');
                }
            },
            "setRange": function(event, range) {
				event.original.preventDefault();
                r.set("range", range);
                drawRange();
            },
            "setLocation": function(latitude, longitude, optZoom, skipRecenter) {
                var map = r.get("map");
                var marker = r.get("marker");
                var location = new google.maps.LatLng(latitude, longitude);
				if(!skipRecenter) {
                map.setCenter(location);
				}
                if (optZoom) {
                    map.setZoom(optZoom);
                }
                if (!marker) {
                    marker = new google.maps.Marker({
                        map: map,
                        position: location,
                        draggable: true,
                        cursor: "move",
                        title: "Drag to choose a specific location",
                    });
                    marker.addListener("dragend", function(e) {
                        r.set("longitude", e.latLng.lng());
                        r.set("latitude", e.latLng.lat());
                        drawRange();
                    });
                    r.set("marker", marker);
                } else {
                    marker.setPosition(location);
                }
                r.set("longitude", longitude);
                r.set("latitude", latitude);
                setStorage("lastLat", latitude);
                setStorage("lastLng", longitude);
                drawRange();
            },
            "myLocation": function() {
                if ("geolocation" in navigator) {
                    navigator.geolocation.getCurrentPosition(function(position) {
                        r.set("search", null);
                        r.set("searchString", null);
                        r.fire("setLocation", position.coords.latitude, position.coords.longitude, 14);
                    }, function(error) {
                        r.set("error", "Sorry, we were unable to get your location: " + error.message);
                    });
                } else {
                    r.set("error", "We cannot get your current location from this browser, sorry.");
                }
            },
            "clearSecondaryMarkers": function() {
                var markers = r.get("markers");
                if (markers) {
                    for (var i = 0; i < markers.length; i++) {
                        markers[i].setMap(null);
                    }
                }

                r.set("markers", []);
            },
            "addSecondaryMarker": function(location, name, hidden) {
                var map = r.get("map");
                if (!map) {
                    return;
                }
                var marker = new google.maps.Marker({
                    map: map,
                    position: location,
                    draggable: false,
                    title: name,
                    icon: secondaryMarker,
                    visible: (!hidden),
                });
                marker.addListener("click", function(e) {
                    var primary = r.get("marker");
                    var markers = r.get("markers");
                    if (!primary || !markers) {
                        return;
                    }

                    for (var i = 0; i < markers.length; i++) {
                        if (markers[i].getPosition().equals(e.latLng)) {
                            markers[i].setVisible(false);
                        } else {
                            markers[i].setVisible(true);
                        }
                    }

                    r.fire("setLocation", e.latLng.lat(), e.latLng.lng(), null, true);
                    //TODO: don't recenter
                });
                r.push("markers", marker);
            },
        });


        function drawRange() {
            if (!r.get("showRange")) {
                return;
            }
            var map = r.get("map");
            var marker = r.get("marker");
            var rangeCircle = r.get("rangeCircle");
            if (!map || !marker) {
                return;
            }
            if (!rangeCircle) {
                rangeCircle = new google.maps.Circle({
                    strokeColor: "#d9534f",
                    strokeOpacity: 0.25,
                    strokeWeight: 2,
                    fillColor: "#d9534f",
                    fillOpacity: 0.25,
                    map: map,
                });
            }
            rangeCircle.setCenter(marker.getPosition());
            rangeCircle.setRadius(r.get("range") * 1609.34); // miles to meters
            r.set("rangeCircle", rangeCircle);
        }

    },
};
