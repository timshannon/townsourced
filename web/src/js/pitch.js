// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */

// Ractive + components
import Ractive from "./ts/ractiveInit";
import Contact from "./components/contact";
import fade from "./lib/ractive-transitions-fade";
import {
    scale
}
from "./lib/ractive-transition-scale";

// ts libs
import {
    err
}
from "./ts/error";

//3rd party
import "./lib/bootstrap/modal";

$(document).ready(function() {
    "use strict";

    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            contact: Contact,
        },
        transitions: {
            fade: fade,
            scale: scale,
        },
        data: function() {
            return {
                slide: {
                    intro: {
                        top: 0,
                        range: 0
                    },
                    problem: {
                        top: 0,
                        range: 0,
                        duration: 10,
                        modScale: 0,
                        searchScale: 0,
                        localScale: 0,
                    },
                    solution: {
                        top: 0,
                        range: 0,
                        duration: 500,
                    },
                    plan: {
                        top: 0,
                        range: 0,
                        duration: 1000,
                    },
                    summary: {
                        top: 0,
                        range: 0,
                        duration: 10,
                    },
                    bio: {
                        top: 0,
                        range: 0,
                        duration: 10,
                    },
                },
                trigger: {
                    timeline: {
                        top: 0,
                        range: 0,
                        duration: 1000,
                        hover: null,
                    },
                },
                contactFade: 0,
                videoTop: 0,
                inModal: false,
            };
        },
        oncomplete: function() {
                this.set("slide.intro.top", 0);
                this.set("slide.problem.top", $("#problem").offset().top);
                this.set("slide.solution.top", $("#solution").offset().top);
                this.set("slide.plan.top", $("#plan").offset().top);
                this.set("slide.summary.top", $("#summary").offset().top);
                this.set("slide.bio.top", $("#bio").offset().top);

                this.set("videoTop", $("#video-position").offset().top);

                if (window.location.hash) {
                    scrollTo(this.get("slide." + window.location.hash.slice(1) + ".top") || 0);
                }
        },
    });

    var slideStops = [
        $("#problem").offset().top,
        $("#solution").offset().top,
        $("#plan").offset().top,
        $("#summary").offset().top,
        $("#bio").offset().top,
    ];

    slideStops.sort(function(a, b) {
        return a - b;
    });
    slideStops.next = function() {
        var top = $(window).scrollTop();
        for (var i = 0; i < this.length; i++) {
            if (this[i] > (top + 1)) {
                return this[i];
            }
        }
        return $(document).height();
    };
    slideStops.prev = function() {
        var top = $(window).scrollTop();
        for (var i = (this.length - 1); i >= 0; i--) {
            if (this[i] < (top - 1)) {
                return this[i];
            }
        }
        return 0;
    };

    r.on({
        "navHover": function(event, kp) {
            if (event.original.type == "mouseout") {
                r.set(kp, false);
            } else {
                r.set(kp, true);
            }
        },
        "scrollTo": function(event, to) {
            event.original.preventDefault();
            scrollTo($(to).offset().top);
        },
        "toTop": function(event) {
            event.original.preventDefault();
            event.original.target.blur();
            event.original.target.parentNode.blur();
            scrollTo(0);
        },
        "nextSlide": function() {
            scrollTo(slideStops.next());
        },
        "prevSlide": function() {
            scrollTo(slideStops.prev());
        },
        "planHover": function(event, item) {
            if (event.original.type == "mouseout") {
                if (r.get("trigger.timeline.hover") == item) {
                    r.set("trigger.timeline.hover", null);
                }
            } else {
                r.set("trigger.timeline.hover", item);
            }
        },
        "contactHover": function(event) {
            if (event.original.type == "mouseout") {
                r.animate("contactFade", 0, {
                    duration: 200
                });
            } else {
                r.animate("contactFade", 1, {
                    duration: 200
                });
            }
        },
        "showContact": function(event) {
            event.original.preventDefault();
            $("#contactModal").modal();
        },
    });

    $("#contactModal").on("shown.bs.modal", function(e) {
        r.set("inModal", true);
    });
    $("#contactModal").on("hidden.bs.modal", function(e) {
        r.set("inModal", false);
    });


    function scrollTo(to) {
        $("html, body").animate({
            scrollTop: to,
        }, 400);
    }

    $(window).on("keyup", function(e) {
        if (r.get("inModal")) {
            return;
        }
        if (e.keyCode == 39 || e.keyCode == 40 || e.keyCode == 32 || e.keyCode == 13 || e.keyCode == 78) {
            e.preventDefault();
            r.fire("nextSlide");
        } else if (e.keyCode == 37 || e.keyCode == 38 || e.keyCode == 80) {
            e.preventDefault();
            r.fire("prevSlide");
        }
    });


    var delay;

    $(window).scroll(function() {
        if (delay) {
            window.clearTimeout(delay);
        }

        delay = window.setTimeout(function() {
            var min = $(window).scrollTop();
            var max = min + $(window).height();
            var set = false;
            $.each(r.get("slide"), function(k, v) {
                var slide = r.get("slide." + k);
                if (!set && slide.top < max && slide.top >= min) {
                    set = true;
                    r.animate("slide." + k + ".range",
                        Math.min((max - slide.top) / (max - min), 1), {
                            duration: slide.duration,
                        });
                } else if (slide.top <= min) {
                    r.set("slide." + k + ".range", 1);
                } else {
                    r.set("slide." + k + ".range", 0);
                }
            });
        }, 10);
    });


    $(window).resize(function() {
        r.set("videoTop", $("#video-position").offset().top);
    });

    var triggerDelay;

    r.observe({
        "slide.problem.range": function(n) {
            r.animate("slide.problem.modScale", n, {
                duration: 500
            });
            r.animate("slide.problem.searchScale", n, {
                duration: 350
            });
            r.animate("slide.problem.localScale", n, {
                duration: 150
            });
        },
        "slide.plan.range": function(n) {
            if (triggerDelay) {
                window.clearTimeout(triggerDelay);
            }

            triggerDelay = window.setTimeout(function() {
                if (n >= 0.9) {
                    r.animate("trigger.timeline.range", 1, {
                        duration: r.get("trigger.timeline.duration"),
                    });
                } else if (n <= 0.7) {
                    r.animate("trigger.timeline.range", 0, {
                        duration: r.get("trigger.timeline.duration"),
                    });
                }
            }, 100);
        },
    });

});
