var gobble = require("gobble");

var path = require('path'),
    fs = require("fs");

module.exports = gobble([
    gobble("src").include("emoji.json").transform(emojiData),
]);


// build maps that can be imported for emoji search and replacement
function emojiData(input) {
    var data = {
        category: {},
        keyword: [],
    };

    input = JSON.parse(input);

    for (var id in input) {
        if (input.hasOwnProperty(id)) {
            var e = input[id];

			if(e.category == "modifier") {
				continue;
				}

            //category
            if (!data.category[e.category]) {
                data.category[e.category] = [];
            }

            data.category[e.category].push({
                shortname: e.shortname,
                img: e.unicode,
            });

            //keyword
            data.keyword.push({
                name: e.name,
                keywords: e.keywords,
                img: e.unicode,
                shortname: e.shortname,
            });
        }
    }

    return "export default " + JSON.stringify(data) + ";";
}
