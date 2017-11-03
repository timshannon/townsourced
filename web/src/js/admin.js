// Townsourced
// Copyright 2016 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
//Ractive
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import ExpandPanel from "./components/expandPanel";
import Chart from "./components/chart";

//ts libs
import {
    htmlPayload,
}
from "./ts/util";

// 3rd party
$(document).ready(function() {
    "use strict";


    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            expandPanel: ExpandPanel,
            chart: Chart,
        },
        data: function() {
            var stats = htmlPayload();
            return {
                stats: stats,
                userCountTrendChart: makeDateCountChart("User Count By Day", stats.userCountTrend),
                townCountTrendChart: makeDateCountChart("Town Count By Day", stats.townCountTrend),
                postCountTrendChart: makeDateCountChart("Post Count By Day", stats.postCountTrend),
            };
        },
    });


    //ractive events
    r.on({

    });

    //functions

    function arrKey(data, key, eachFunc) {
        var result = [];
        for (var i = 0; i < data.length; i++) {
            if (eachFunc) {
                result.push(eachFunc(data[i][key]));
            } else {
                result.push(data[i][key]);
            }
        }

        return result;
    }

    function makeDateCountChart(label, data) {
        return {
            type: "line",
            data: {
                labels: arrKey(data, "date"),
                datasets: [{
                    label: label,
                    fill: false,
                    backgroundColor: "#83bf48",
                    borderColor: "#246b08",
                    data: arrKey(data, "count"),
                }],
            },
            options: {
                responsive: true,
                scales: {
                    xAxes: [{
                        type: "time",
                        time: {
                            unit: "week",
                            tooltipFormat: "MMMM Do YYYY",
                        },
                    }],
                },
            },
        };
    }

});
