// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/* jshint  esnext: true, strict: true */
import csrf from "./csrf";

export
function contact(email, subject, message) {
    "use strict";
    return csrf.ajax({
        type: "POST",
        url: "/api/v1/contact/",
        data: JSON.stringify({
            email: email,
            subject: subject,
            message: message,
        }),
    });
}
