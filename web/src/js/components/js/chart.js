// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.
/* jshint  esnext: true*/

import Ractive from "ractive";

//3rd party
import "../../lib/Chart.bundle.js";

export
default {
    decorators: {
        chart: function(node) {
            var r = this;
            var chart = new Chart(node, r.get("chart"));
            return {
                teardown: function() {
                    chart.destroy();
                },
            };
        },
    },
    isolated: true,
    data: function() {
        return {
            chart: {},
        };
    },
    oninit: function() {
        // global defaults for charts
        Chart.defaults.global.defaultFontColor = "#333";
        Chart.defaults.global.defaultFontFamily = "'Droid Sans', 'Helvetica Neue', 'Helvetica', 'Arial', 'sans-serif'";

        //line
        /*Chart.defaults.global.elements.line.backgroundColor = "transparent";*/
        Chart.defaults.global.elements.line.borderColor = "#83bf48";

        //point
        Chart.defaults.global.elements.point.backgroundColor = "#246b08";
    },
    onrender: function() {
        var r = this;

        r.on({});
    },
};
