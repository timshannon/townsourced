// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true */
import csrf from "./csrf";

export
function upload(file, progress) {
        var formData;

        var options = {
            cache: false,
            processData: false,
            contentType: false,
            xhr: function() {
                var xhr = new window.XMLHttpRequest();
                xhr.upload.addEventListener("progress", progress, false);
                return xhr;
            },
        };

        if (file instanceof FormData) {
            options.data = file;
        } else if (file instanceof File) {
            options.data = new FormData();

            options.data.append(file.name,
                file, file.name);
        } else {
            //not supported
            throw "FileData type not supported";
        }

        return csrf.ajax("POST", "/api/v1/image/", options);
    }
