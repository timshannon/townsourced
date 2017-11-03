var gobble = require("gobble");

var path = require('path'),
    sander = require("sander"),
    fs = require("fs"),
    rcu = require('rcu'),
    builders = require('rcu-builders'),
    juice = require('juice');

var jsRoots = [
    "page",
    "root",
    "search",
    "searchLocation",
    "public",
    "demo",
    "3rdparty",
    "user",
    "newtown",
    "town",
    "townSettings",
    "townSearch",
    "editpost",
    "post",
    "message",
    "forgotPassword",
    "pitch",
    "unauthorized",
    "help",
    "admin",
    "welcome",
	"bookmarklet",
];

module.exports = gobble([
    gobble("src"),
    css(gobble("src/css"), "townsourced.less", "townsourced.css").moveTo("css"),
    htmlCSS("src"),
    js(),
    email("src/email").moveTo("email"),
]);

function css(g, fin, fout) {
    fout = fout || fin;

    return g.transform("less", {
        src: fin,
        dest: fout,
        compress: true,
    }).transform("autoprefixer", {
        src: fin,
        dest: fout,
        browsers: [
            "Android 2.3",
            "Android >= 4",
            "Chrome >= 20",
            "Firefox >= 24",
            "Explorer >= 8",
            "iOS >= 6",
            "Opera >= 12",
            "Safari >= 6"
        ]
    });
}

function htmlCSS(dir) {
    var files = fs.readdirSync(dir);

    var g = [];
    for (var i = 0; i < files.length; i++) {
        if (path.extname(files[i]) == ".html") {
            g = g.concat(css(gobble(dir), files[i]));
        }
    }
    return g;
}

function js() {
    var jsFiles = [];

    var jsSrc = gobble("src/js").transform(processAllComponents, {
        type: "es6",
        accept: '.html',
        ext: '.js',
    });


    for (i = 0; i < jsRoots.length; i++) {
        jsFiles.push(jsProcess(jsSrc, jsRoots[i]));
    }

    return jsFiles;
}


function jsProcess(g, file) {
    if (gobble.env() != "production" && gobble.env() != "development" && gobble.env() != file) {
        //here's where I abuse gobble.env() to do really quick builds on a single page
        // gobble watch static -e page
        return g;
    }
    file = file + ".js";
    g = g.transform("rollup", {
        entry: file,
        format: "umd",
        external: ["ractive"],
    });

    if (gobble.env() === "production") {
        g = g.transform('uglifyjs');
    }
    return g.moveTo("js");
}


//TODO: this mess below can be mostly thrown away if filetransformers could handle async callbacks
rcu.init(require('ractive'));

function processAllComponents(inputdir, outputdir, options, callback) {
    var files = sander.lsrSync(inputdir);
    var promises = [];

    for (var i = 0; i < files.length; i++) {
        if (path.extname(files[i]) == options.accept) {
            var fin = path.join(inputdir, files[i]);
            var fout = path.join(outputdir, files[i].replace(path.extname(files[i]), options.ext));
            promises.push(processComponent(fin, fout, options));
        } else {
            promises.push(sander.copyFile(inputdir, files[i]).to(outputdir, files[i]));
        }
    }

    sander.Promise.all(promises).then(function() {
        callback();
    }, callback);
}

function processComponent(fin, fout, options) {
    var builder = builders[options.type || 'amd'];

    if (!builder) {
        throw new Error('Cannot convert Ractive component to "' + options.type + '". Supported types: ' + Object.keys(builders));
    }
    var source = sander.readFileSync(fin).toString();

    options.filename = fin;

    options.sourceMap = options.sourceMap !== false;
    if (options.sourceMap) {
        options.sourceMapFile = fout;
        options.sourceMapSource = fin;
    }

    var parsed = rcu.parse(source);

    return new sander.Promise(function(resolve, reject) {
        require("less").render(parsed.css, options, function(error, result) {
            if (error) {
                reject("Error parsing less css for ractive component: " + error);
            }

            parsed.css = result.css;
            sander.writeFile(fout, builder(parsed, options))
                .then(function() {
                    resolve();
                });
        });
    });

}

function email(dir) {
    var files = fs.readdirSync(dir);

    var g = [];
    for (var i = 0; i < files.length; i++) {
        if (files[i].indexOf(".template.html") !== -1) {
            g = g.concat(gobble(dir).transform("concat", {
                dest: files[i],
                files: ["header.part.html", files[i], "footer.part.html"],
            }).transform("less", {
                src: files[i],
                dest: files[i],
                compress: true,
            }).transform(juice));
        }
    }
    return gobble(g);
}
