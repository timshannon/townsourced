// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.
/* jshint  esnext: true, strict: true */
//Ractive
import Ractive from "./ts/ractiveInit";
import Page from "./components/page";
import Modal from "./components/modal";
import ImageUpload from "./components/imageUpload";
import Color from "./components/color";
import Map from "./components/map";
import UserIcon from "./components/userIcon";
import Alert from "./components/alert";
import PreviewEditor from "./components/previewEditor";
import Header from "./components/header";
import UserSelect from "./components/userSelect";

//ts libs
import * as Town from "./ts/town";
import {
    categories
}
from "./ts/post";
import {
    get as userGet
}
from "./ts/user";
import {
    htmlPayload,
    isEmail,
    escapeRegExp,
    since,
}
from "./ts/util";
import {
    err
}
from "./ts/error";

import * as Storage from "./ts/storage";

// 3rd party
import "./lib/bootstrap/tab";

$(document).ready(function() {
    "use strict";


    var r = new Ractive({
        el: "body",
        template: "#tMain",
        components: {
            page: Page,
            modal: Modal,
            imageUpload: ImageUpload,
            color: Color,
            map: Map,
            userIcon: UserIcon,
            alert: Alert,
            previewEditor: PreviewEditor,
            header: Header,
            userSelect: UserSelect,
        },
        data: function() {
            return {
                error: null,
                inviteUser: null,
                userBan: null,
                town: htmlPayload(),
                users: {},
                categories: categories,
                autoModAdvanced: Storage.get("autoModAdvanced"),
                headerHeight: Town.headerHeight,
                headerWidth: Town.headerWidth,
                maxTownDescription: Town.maxTownDescription,
                maxTownName: Town.maxTownName,
                inviteMemberEmail: "",
                inviteMemberEmailErr: null,
                inviteMemberErr: null,
                inviteByUsername: true,
                memberSortType: "date",
                since: since,
                invitesOriginal: [],
            };
        },
    });

    setTown(r.get("town"));

    //ractive events
    r.on({
        "navbar.userLoaded": function(currentUser) {
            if (currentUser) {
                r.set("currentUser", currentUser);

                currentUser.self = true;
                r.set("users." + currentUser.username, currentUser);
                return;
            }

        },
        "back": function() {
            window.location = "/town/" + r.get("town.key");
        },
        "header.imageModal": function(event) {
            $("#imageModal").modal();
            r.findComponent("imageUpload").fire("reset");
        },
        "imageSet": function(event) {
            event.original.preventDefault();
            var img = r.get("newHeaderImage");
            Town.setHeaderImage(r.get("town.key"), img.key, img.x0, img.y0, img.x1, img.y1, r.get("town.vertag"))
                .done(function() {
                    location.reload();
                })
                .fail(function(result) {
                    r.findComponent("imageUpload").set("error", err(result).message);
                });
        },
        "removeImage": function() {
            Town.removeHeaderImage(r.get("town.key"), r.get("town.vertag"))
                .done(function() {
                    location.reload();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "color.setColor": function(event, color) {
            Town.setColor(r.get("town.key"), color, r.get("town.vertag"))
                .done(function() {
                    r.set("colorErr", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("colorErr", err(result).message);
                });
        },
        "nameModal": function(event) {
            event.original.preventDefault();
            $("#nameModal").on("shown.bs.modal", function() {
                $("#nameChange").focus();
            });

            r.set("nameChange", r.get("town.name"));

            r.set("nameErr", false);

            $("#nameModal").modal();
        },
        "nameSave": function(event) {
            event.original.preventDefault();
            Town.setName(r.get("town.key"), event.context.nameChange, r.get("town.vertag"))
                .done(function() {
                    loadTown();
                    $("#nameModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("nameErr", err(result).message);
                });
        },
        "descriptionModal": function(event) {
            event.original.preventDefault();
            $("#descriptionModal").on("shown.bs.modal", function() {
                $("#descriptionChange").focus();
            });

            r.set("descriptionChange", r.get("town.description"));

            r.set("descriptionErr", null);

            $("#descriptionModal").modal();
        },
        "descriptionSave": function(event) {
            event.original.preventDefault();
            Town.setDescription(r.get("town.key"), event.context.descriptionChange, r.get("town.vertag"))
                .done(function() {
                    loadTown();
                    $("#descriptionModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("descriptionErr", err(result).message);
                });
        },
        "inviteModModal": function() {
            $("#inviteModModal").on("shown.bs.modal", function() {
                $("#inviteUsername").focus();
            });

            r.set("inviteUser", null);
            r.set("inviteErr", null);
            $("#inviteModModal").modal();
        },
        "sendInvite": function(event) {
            event.original.preventDefault();
            if (!r.get("inviteUser.username")) {
                r.set("inviteErr", "Please select a user to invite");
                return;
            }
            Town.inviteMod(r.get("town.key"), r.get("inviteUser.username"))
                .done(function() {
                    loadTown();
                    $("#inviteModModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("inviteErr", err(result).message);
                });
        },
        "resignAdmin": function() {
            $("#resignModal").on("shown.bs.modal", function() {
                $("#resignCheck").focus();
            });

            r.set("resignCheck", null);

            r.set("resignErr", false);

            $("#resignModal").modal();

        },
        "resignConfirm": function(event) {
            event.original.preventDefault();
            if (r.get("resignCheck") !== r.get("currentUser.username")) {
                r.set("resignErr", "Incorrect username");
                return;
            }
            Town.removeMod(r.get("town.key"))
                .done(function() {
                    r.fire("back");
                })
                .fail(function(result) {
                    r.set("resignErr", err(result).message);
                });
        },
        "setPrivacy": function(event) {
            r.toggle("town.private");
            Town.setPrivacy(r.get("town.key"), r.get("town.private"), r.get("town.vertag"))
                .done(function() {
                    loadTown();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "inviteMemberModal": function() {
            $("#inviteMemberModal").on("shown.bs.modal", function() {
                $("#inviteMemberName").focus();
            });

            r.set("inviteMember", null);
            r.set("inviteMemberErr", null);
            r.set("inviteMemberEmail", "");
            $("#inviteMemberModal").modal();
        },
        "sendMemberInvite": function(event) {
            event.original.preventDefault();
            if (!r.get("inviteMember") && !r.get("inviteMemberEmail")) {
                r.set("inviteMemberErr", "Please select a user to invite or specify an email address");
                return;
            }

            if (r.get("inviteMemberEmail")) {
                if (!isEmail(r.get("inviteMemberEmail"))) {
                    r.set("inviteMemberErr", "Invalid Email address");
                    return;
                }
            }

            Town.invite(r.get("town.key"), r.get("inviteMember.username"), r.get("inviteMemberEmail"))
                .done(function() {
                    loadTown();
                    $("#inviteMemberModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("inviteMemberErr", err(result).message);
                });
        },
        "removeMemberInvite": function(event) {
            Town.removeInvite(r.get("town.key"), event.context.invite.username)
                .done(function() {
                    loadTown();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "validateInviteEmail": function(event) {
            r.set("inviteMemberErr", null);
            if (r.get("inviteMemberEmail")) {
                if (!isEmail(r.get("inviteMemberEmail"))) {
                    r.set("inviteMemberErr", "Invalid Email address");
                    return;
                }
            }
        },
        "toggleInviteType": function(event) {
            event.original.preventDefault();
            r.set("inviteMemberErr", null);
            if (r.get("inviteByUsername")) {
                r.set("inviteByUsername", false);
                r.set("inviteMember", null);
                return;
            }
            r.set("inviteByUsername", true);
            r.set("inviteMemberEmail", null);
            return;
        },
        "infoModal": function(event) {
            event.original.preventDefault();
            r.set("infoErr", null);
            r.set("infoChange", r.get("town.information"));
            r.set("atTowns", [r.get("town.key")]);
            r.set("atUsers", r.get("users"));
            $("#infoModal").modal();
        },
        "infoHide": function() {
            r.set("infoErr", null);
        },
        "infoSave": function() {
            Town.setInformation(r.get("town.key"), r.get("infoChange"), r.get("town.vertag"))
                .done(function() {
                    loadTown();
                    $("#infoModal").modal("hide");
                })
                .fail(function(result) {
                    r.set("infoErr", err(result).message);
                });
        },
        "autoModCategoryModal": function() {
            $("#autoModCategoryModal").modal();
        },
        "autoModCategoryModalHide": function() {
            r.set("autoModCategoryErr", null);
        },
        "autoModUserModal": function() {
            r.set("minUserDaysValue", r.get("town.autoModerator.minUserDays"));
            $("#autoModUserModal").modal();
        },
        "autoModUserModalHide": function() {
            r.set("autoModUserErr", null);
            r.set("userDaysErr", null);
            r.set("banUserErr", null);
        },
        "autoModContentModal": function() {
            r.set("maxNumLinksValue", r.get("town.autoModerator.maxNumLinks"));
            $("#autoModContentModal").modal();
        },
        "autoModContentModalHide": function() {
            r.set("autoModContentErr", null);
            r.set("maxLinksErr", null);
            r.set("autoModContentErr", null);
            r.set("regexpBan", null);
            r.set("regexpReason", null);
            r.set("regexpErr", null);
            r.set("reasonErr", null);
        },
        "addCategory": function(event, category) {
            Town.addAutoModCategory(r.get("town.key"), category)
                .done(function() {
                    loadTown();
                    r.set("autoModCategoryErr", null);
                })
                .fail(function(result) {
                    r.set("autoModCategoryErr", err(result).message);
                });
        },
        "removeCategory": function(event, category) {
            Town.removeAutoModCategory(r.get("town.key"), category)
                .done(function() {
                    r.set("autoModCategoryErr", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("autoModCategoryErr", err(result).message);
                });

        },
        "setMinUserDays": function(event) {
            event.original.preventDefault();
            Town.setAutoModMinUserDays(r.get("town.key"), r.get("minUserDaysValue"))
                .done(function() {
                    loadTown();
                    r.set("userDaysErr", null);
                })
                .fail(function(result) {
                    r.set("userDaysErr", err(result).message);
                });
        },
        "banUser": function(user) {
            Town.addAutoModUser(r.get("town.key"), user.username)
                .done(function() {
                    r.set("banUserErr", null);
                    r.set("userBan", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("banUserErr", err(result).message);
                });

        },
        "removeBan": function(event) {
            Town.removeAutoModUser(r.get("town.key"), event.context)
                .done(function() {
                    r.set("autoModUserErr", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("autoModUserErr", err(result).message);
                });
        },
        "setMaxNumLinks": function(event) {
            event.original.preventDefault();
            Town.setAutoModMaxNumLinks(r.get("town.key"), r.get("maxNumLinksValue"))
                .done(function() {
                    loadTown();
                    r.set("maxLinksErr", null);
                })
                .fail(function(result) {
                    r.set("maxLinksErr", err(result).message);
                });
        },
        "autoModAdvanced": function() {
            r.toggle("autoModAdvanced");
            r.set("autoModContentErr", null);
            r.set("regexpBan", null);
            r.set("regexpReason", null);
            r.set("regexpErr", null);
            r.set("reasonErr", null);
            Storage.set("autoModAdvanced", r.get("autoModAdvanced"));
        },
        "addWordBan": function(event) {
            event.original.preventDefault();
            if (!r.get("regexpBan") || !r.get("regexpBan").trim()) {
                return;
            }
            var regexp = "(?i)\\b" + escapeRegExp(r.get("regexpBan").trim()) + "\\b";
            var reason = "The phrase " + r.get("regexpBan") + " is not allowed in posts to this town.";

            Town.addAutoModRegexp(r.get("town.key"), regexp, reason)
                .done(function() {
                    r.set("regexpBan", null);
                    r.set("regexpReason", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("autoModContentErr", err(result).message);
                });
        },
        "addRegexpBan": function(event) {
            event.original.preventDefault();
            r.set("regexpErr", null);
            r.set("reasonErr", null);

            if (!r.get("regexpBan") || !r.get("regexpBan").trim()) {
                r.set("regexpErr", "An expression is required");
                return;
            }
            if (!r.get("regexpReason") || !r.get("regexpReason").trim()) {
                r.set("reasonErr", "An reason is required");
                return;
            }

            Town.addAutoModRegexp(r.get("town.key"), r.get("regexpBan"), r.get("regexpReason"))
                .done(function() {
                    r.set("regexpBan", null);
                    r.set("regexpReason", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("autoModContentErr", err(result).message);
                })
                .always(function() {
                    r.set("regexpErr", null);
                    r.set("reasonErr", null);
                });
        },
        "removeExpression": function(event) {
            Town.removeAutoModRegexp(r.get("town.key"), event.context.regexp)
                .done(function() {
                    r.set("autoModContentErr", null);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("autoModContentErr", err(result).message);
                });
        },
        "acceptInviteRequest": function(event) {
            event.original.preventDefault();
            var username = event.context.who.username;
            Town.acceptInviteRequest(r.get("town.key"), username)
                .done(function() {
                    // remove user from memory, so it's reloaded with new membership
                    r.set("users." + username, null);
                    r.set("town.inviteRequests", []);
                    loadTown();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "rejectInviteRequest": function(event) {
            event.original.preventDefault();
            Town.rejectInviteRequest(r.get("town.key"), event.context.who.username)
                .done(function() {
                    loadTown();
                })
                .fail(function(result) {
                    r.set("error", err(result).message);
                });
        },
        "memberSort": function(event, type) {
            event.original.preventDefault();
            r.set("memberSortType", type);
            sortInvites(type);
        },
    });

    r.observe({
        "town.autoModerator.minUserDays": function(newval) {
            r.set("minUserDaysValue", newval);
        },
        "town.autoModerator.maxNumLinks": function(newval) {
            r.set("maxNumLinksValue", newval);
        },

    });

    //functions
    function loadTown() {
        Town.get(r.get("town.key"))
            .done(function(result) {
                setTown(result.data);
            })
            .fail(function(result) {
                r.set("error", err(result, true).message);
            });
    }

    function setTown(town) {

        r.set("town", town);
        document.title = town.name + " - settings - townsourced";
        Town.setTheme(town.color);

        r.set("infoParsed", ts.processMessage(town.information));

        if (town.moderators) {
            for (var i = 0; i < town.moderators.length; i++) {
                loadUser(town.moderators[i].username);
            }
        }

        if (town.invites) {
            r.set("invitesOriginal", town.invites.slice(0));
            for (var j = 0; j < town.invites.length; j++) {
                loadUser(town.invites[j]);
            }

            sortInvites(r.get("memberSortType"));
        }

        if (town.inviteRequests) {
            for (var k = 0; k < town.inviteRequests.length; k++) {
                loadUser(town.inviteRequests[k].who);
            }
        }


    }

    function loadUser(username) {
        if (r.get("users." + username)) {
            return;
        }

        r.set("users." + username, {
            username: username
        });

        userGet(username)
            .done(function(result) {
                var user = result.data;
                user.inactive = !Town.isModerator(r.get("town"), user);
                user.member = Town.isMember(r.get("town"), user);
                if (r.get("currentUser") && r.get("currentUser.username") == user.username) {
                    user.self = true;
                }
                r.set("users." + user.username, user);
            });
    }

    function sortInvites(type) {
        var invites = r.get("town.invites");
        if (!invites) {
            return;
        }
        if (type == "status") {
            invites.sort(function(a, b) {
                var userA = r.get("users." + a);
                var userB = r.get("users." + b);
                if (userB.member && (!userA.member)) {
                    return 1;
                }
                if (userA.member && (!userB.member)) {
                    return -1;
                }
                return 0;
            });

        } else if (type == "name") {
            invites.sort(function(a, b) {
                var nameA = r.get("users." + a + ".name") || a;
                var nameB = r.get("users." + b + ".name") || b;

                if (nameA > nameB) {
                    return 1;
                }
                if (nameA < nameB) {
                    return -1;
                }
                // a must be equal to b
                return 0;
            });
        } else {
            r.set("town.invites", r.get("invitesOriginal").slice(0));
        }
    }


});
