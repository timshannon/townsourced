// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true*/

import Ractive from "ractive";
import {
    newMap,
    secondaryMarker,
    haversine,
}
from "../../ts/map";

import {
    search as townSearch
}
from "../../ts/town";

import {
    err
}
from "../../ts/error";

export
default {
    decorators: {
        map: function(node) {
            var r = this;

            var map = newMap(node, true);
            var placeSearch = new google.maps.places.PlacesService(map);

            var delay;

            map.addListener("bounds_changed", function(e) {
                if (delay) {
                    window.clearTimeout(delay);
                }
                delay = window.setTimeout(function() {
                    r.fire("loadTowns");
                }, 500);
            });

            r.set("map", map);
            r.set("placeSearch", placeSearch);

            return {
                teardown: function() {
                    r.set("map", undefined);
                    r.set("placeSearch", undefined);
                },
            };
        },
    },
    isolated: true,
    data: function() {
        return {
            showMap: false,
            loading: false,
            towns: [],
            markers: {},
            town: null,
            pickTown: false,
        };
    },
    onrender: function() {
        var r = this;

        r.on({
            "search": function(searchValue) {
                r.set("showMap", true);
                r.set("error", null);
                if (!searchValue) {
                    return;
                }

                var placeSearch = r.get("placeSearch");
                var map = r.get("map");

                if (!placeSearch || !map) {
                    return;
                }

                placeSearch.textSearch({
                    bounds: map.getBounds(),
                    query: searchValue,
                }, function(places) {
                    if (places.length === 0) {
                        return;
                    }

                    var bounds = new google.maps.LatLngBounds();
                    places.forEach(function(place) {
                        if (place.geometry.viewport) {
                            // Only geocodes have viewport.
                            bounds.union(place.geometry.viewport);
                        } else {
                            bounds.extend(place.geometry.location);
                        }
                    });

                    r.fire("setLocation", places[0].geometry.location.lat(), places[0].geometry.location.lng(), bounds);
                    r.fire("loadTowns", 1);
                });
            },
            "reset": function() {
                r.set("showMap", false);
                r.set("error", null);
            },
            "setLocation": function(latitude, longitude, mapBounds) {
                r.set("error", null);
                if (!latitude || !longitude) {
                    return;
                }
                r.set("showMap", true).then(function() {
                    var map = r.get("map");
                    var location = new google.maps.LatLng(latitude, longitude);
                    map.setCenter(location);
                    if (mapBounds) {
                        map.fitBounds(mapBounds);
                    } else {
                        map.setZoom(12);
                    }
                });
            },
            "pickTownAtLocation": function(latitude, longitude, mapBounds) {
                r.set("pickTown", true);
                r.fire("setLocation", latitude, longitude, mapBounds);
            },
            "resetMarkers": function() {
                r.set("error", null);
                $.each(r.get("markers"), function(k, v) {
                    v.setMap(null);
                });
                r.set("markers", {});
                setMarkers();
            },
            "selectTown": function(town) {
                r.fire("resetMarkers");

                r.set("town", town);
                r.fire("setLocation", town.location.Lat, town.location.Lon);
            },
            "myLocation": function() {
                if ("geolocation" in navigator) {
                    navigator.geolocation.getCurrentPosition(function(position) {
                        r.fire("resetMarkers");
                        r.set("pickTown", true);
                        r.fire("setLocation", position.coords.latitude, position.coords.longitude);

                    }, function(error) {
                        return;
                    });
                } else {
                    r.set("error", "We cannot get your current location from this browser, sorry.");
                }
            },
            "loadTowns": function() {
                var map = r.get("map");
                if (map) {
                    getTowns(map.getBounds());
                }
            },
        });

        function getTowns(bounds) {
            r.set("loading", true);

            townSearch({
                    northBounds: bounds.getNorthEast().lat(),
                    southBounds: bounds.getSouthWest().lat(),
                    eastBounds: bounds.getNorthEast().lng(),
                    westBounds: bounds.getSouthWest().lng(),
                })
                .done(function(result) {
                    r.set("towns", result.data);
                    if (r.get("pickTown")) {
                        if (result.data && result.data.length && result.data.length > 0) {
                            r.set("town", result.data[0]);
                            r.fire("townSet", result.data[0]);
                        }
                        r.set("pickTown", false);
                    }
                    r.set("loading", false);
                    setMarkers();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                    r.set("loading", false);
                });

        }

        function setMarkers() {
            var towns = r.get("towns");

            if (!towns || !towns.length || towns.length === 0) {
                return;
            }
            var markers = r.get("markers");

            for (var i = 0; i < towns.length; i++) {
                if (!markers[towns[i].key]) {
                    addMarker(towns[i]);
                }
            }

        }

        function addMarker(town) {
            var map = r.get("map");
            if (!map) {
                return;
            }

            r.set("markers." + town.key, {});

            window.setTimeout(function() {
                var icon;
                var primary = r.get("town");
                if (primary && primary.key === town.key) {
                    icon = null;
                } else {
                    icon = secondaryMarker;
                }
                var marker = new google.maps.Marker({
                    map: map,
                    position: {
                        lat: town.location.Lat,
                        lng: town.location.Lon,
                    },
                    title: town.name,
                    animation: google.maps.Animation.DROP,
                    clickable: true,
                    icon: icon,
                });
                marker.addListener("click", function(e) {
                    selectTownFromLocation(e.latLng);
                });
                r.set("markers." + town.key, marker);

            }, Math.random() * (300 - 100) + 100);
        }

        function selectTownFromLocation(location) {
            var towns = r.get("towns");
            if (!towns || !towns.length || towns.length === 0) {
                return;
            }

            var nearest = -1;
            var index = 0;

            for (var i = 0; i < towns.length; i++) {
                var dist = haversine(location.lat(), location.lng(), towns[i].location.Lat, towns[i].location.Lon);

                if (dist < nearest || nearest == -1) {
                    nearest = dist;
                    index = i;
                }
            }

            r.set("town", towns[index]);
            r.fire("townSet", towns[index]);

            var markers = r.get("markers");
            if (!markers) {
                return;
            }

            $.each(markers, function(k, v) {
                if (v.getPosition().equals(location)) {
                    v.setIcon(null);
                } else {
                    v.setIcon(secondaryMarker);
                }
            });
        }
    },
};
