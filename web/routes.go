// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"
	"path/filepath"
	"strings"

	"git.townsourced.com/townsourced/httprouter"
	"github.com/timshannon/townsourced/app"
)

// TODO: CSP report-uri support

type subDomainHandler map[string]http.Handler

func (s subDomainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.Host, ".")

	sub := parts[0]

	if devMode && subDomainForce != "" {
		sub = subDomainForce
	} else if len(parts) < 3 {
		sub = "www"
		if r.Host != canonicalHost {
			// no subdomain, redirect to canonical host
			http.Redirect(w, r, siteURL(r, r.URL.Path).String(), http.StatusMovedPermanently)
			return
		}
	}

	if handler := s[sub]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		four04(w, r)
	}
}

func setRoutes() subDomainHandler {

	sd := make(subDomainHandler)

	sd["www"] = wwwRoutes()

	return sd
}

func wwwRoutes() http.Handler {
	static := "./web/static"

	four04Handler = newStaticHandler(filepath.Join(static, "404.html"), true)
	errorHandler = newStaticHandler(filepath.Join(static, "error.html"), true)
	unauthorizedHandler = newStaticHandler(filepath.Join(static, "unauthorized.html"), true)

	rootHandler := &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		NotFound:               http.HandlerFunc(four04),
		MethodNotAllowed:       http.HandlerFunc(four04),
		PanicHandler:           panicHandler,
	}

	//static dirs
	serveStaticDir(rootHandler, "/fonts/", filepath.Join(static, "fonts"), true)
	serveStaticDir(rootHandler, "/js/", filepath.Join(static, "js"), true)
	serveStaticDir(rootHandler, "/css/", filepath.Join(static, "css"), true)
	serveStaticDir(rootHandler, "/images/", filepath.Join(static, "images"), false)

	// large files
	serveStaticDir(rootHandler, "/bin/", "./web/bin", false)

	//static pages
	rootHandler.Handler("GET", "/error/", newStaticHandler(filepath.Join(static, "error.html"), true))
	rootHandler.Handler("GET", "/404/", four04Handler)
	rootHandler.Handler("GET", "/429/", newStaticHandler(filepath.Join(static, "429.html"), true))
	rootHandler.Handler("GET", "/3rdparty/", newStaticHandler(filepath.Join(static, "3rdparty.html"), true))
	rootHandler.Handler("GET", "/attribution/", newStaticHandler(filepath.Join(static, "attribution.html"), true))

	// *** Templates

	//root
	rootHandler.GET("/", templateHandler{
		handler: rootTemplate,
		templateFiles: []string{
			filepath.Join(static, "root.template.html"),
			filepath.Join(static, "public.template.html"),
			filepath.Join(static, "demo.template.html"),
		},
	}.ServeHTTP)

	//user
	userTemplateHandler := templateHandler{
		handler:       userTemplate,
		templateFiles: []string{filepath.Join(static, "user.template.html")},
	}

	rootHandler.GET("/user/", userTemplateHandler.ServeHTTP)
	rootHandler.GET("/user/:user", userTemplateHandler.ServeHTTP)

	// user welcome
	rootHandler.GET("/welcome/", templateHandler{
		handler:       welcomeTemplate,
		templateFiles: []string{filepath.Join(static, "welcome.template.html")},
	}.ServeHTTP)

	//post
	postTemplatePageHandler := templateHandler{
		handler:       postTemplate,
		templateFiles: []string{filepath.Join(static, "post.template.html")},
	}

	rootHandler.GET("/post/:post", postTemplatePageHandler.ServeHTTP)
	rootHandler.GET("/post/:post/comment/:comment", postTemplatePageHandler.ServeHTTP)

	editPostTemplateHandler := templateHandler{
		handler:       editPostTemplate,
		templateFiles: []string{filepath.Join(static, "editpost.template.html")},
	}

	rootHandler.GET("/editpost/:post", editPostTemplateHandler.ServeHTTP)
	rootHandler.GET("/newpost/", editPostTemplateHandler.ServeHTTP)

	rootHandler.GET("/share/", templateHandler{
		handler:       shareTemplate,
		templateFiles: []string{filepath.Join(static, "editpost.template.html")},
	}.ServeHTTP)

	//town
	rootHandler.GET("/newtown", templateHandler{
		handler:       newtownTemplate,
		templateFiles: []string{filepath.Join(static, "newtown.template.html")},
	}.ServeHTTP)

	rootHandler.GET("/town/:town", templateHandler{
		handler:       townTemplate,
		templateFiles: []string{filepath.Join(static, "town.template.html")},
	}.ServeHTTP)
	rootHandler.GET("/town/:town/settings/", templateHandler{
		handler:       townSettingsTemplate,
		templateFiles: []string{filepath.Join(static, "townSettings.template.html")},
	}.ServeHTTP)

	rootHandler.GET("/town/", templateHandler{
		handler:       townsearchTemplate,
		templateFiles: []string{filepath.Join(static, "townSearch.template.html")},
	}.ServeHTTP)

	//search
	searchTemplateHandler := templateHandler{
		handler:       searchTemplate,
		templateFiles: []string{filepath.Join(static, "search.template.html")},
	}
	rootHandler.GET("/search", searchTemplateHandler.ServeHTTP)
	rootHandler.GET("/town/:town/search", searchTemplateHandler.ServeHTTP) // search on a specific town

	rootHandler.GET("/search/location", templateHandler{ //location based search
		handler:       searchLocationTemplate,
		templateFiles: []string{filepath.Join(static, "searchLocation.template.html")},
	}.ServeHTTP)

	//forgot password
	rootHandler.GET("/"+app.UserForgotPasswordPath+"/:token", templateHandler{
		handler:       forgotPasswordTemplate,
		templateFiles: []string{filepath.Join(static, "forgotPassword.template.html")},
	}.ServeHTTP)

	//confirm Email
	rootHandler.GET("/"+app.UserConfirmEmailPath+"/:token", templateHandler{
		handler:       confirmEmailTemplate,
		templateFiles: []string{filepath.Join(static, "confirmEmail.template.html")},
	}.ServeHTTP)

	//pitch
	rootHandler.Handler("GET", "/pitch/", &pitchHandler{
		standard: newStaticHandler(filepath.Join(static, "pitch.html"), true),
		simple:   newStaticHandler(filepath.Join(static, "pitch_simple.html"), true),
	})

	//help
	helpTemplateHandler := templateHandler{
		handler:       helpTemplate,
		templateFiles: []string{filepath.Join(static, "help.template.html")},
	}

	rootHandler.GET("/help/", helpTemplateHandler.ServeHTTP)
	rootHandler.GET("/help/:key", helpTemplateHandler.ServeHTTP)

	//admin dashboard
	rootHandler.GET("/admin/", templateHandler{
		handler:       adminTemplate,
		templateFiles: []string{filepath.Join(static, "admin.template.html")},
	}.ServeHTTP)

	//API

	//town
	rootHandler.GET("/api/v1/town/", makeHandle(townSearch))
	rootHandler.POST("/api/v1/town/", makeHandle(townPostNew))

	//	specific town
	rootHandler.GET("/api/v1/town/:town/", makeHandle(townGet))
	rootHandler.PUT("/api/v1/town/:town/", makeHandle(townPut))
	// 	image
	rootHandler.PUT("/api/v1/town/:town/image/", makeHandle(townPutImage))
	rootHandler.DELETE("/api/v1/town/:town/image/", makeHandle(townDeleteImage))
	// 	moderators
	rootHandler.POST("/api/v1/town/:town/moderator/", makeHandle(townPostMod))     //invite new mod
	rootHandler.PUT("/api/v1/town/:town/moderator/", makeHandle(townPutMod))       //accept mod invite
	rootHandler.DELETE("/api/v1/town/:town/moderator/", makeHandle(townDeleteMod)) // resign mod
	// 	private town invites
	rootHandler.POST("/api/v1/town/:town/invite/", makeHandle(townPostInvite))
	rootHandler.DELETE("/api/v1/town/:town/invite/", makeHandle(townDeleteInvite))
	rootHandler.GET("/"+app.TownInviteAcceptPath+"/:token", makeHandle(townAcceptInvite))
	//	private town invite requests
	rootHandler.POST("/api/v1/town/:town/invite/request/", makeHandle(townPostInviteRequest))     // request a private town invite
	rootHandler.PUT("/api/v1/town/:town/invite/request/", makeHandle(townPutInviteRequest))       // accept a request
	rootHandler.DELETE("/api/v1/town/:town/invite/request/", makeHandle(townDeleteInviteRequest)) // reject a request

	//	automod category
	rootHandler.POST("/api/v1/town/:town/automod/category", makeHandle(townPostAutoModCategory))
	rootHandler.DELETE("/api/v1/town/:town/automod/category", makeHandle(townDeleteAutoModCategory))
	//	automod minUserDays, maxNumLinks
	rootHandler.PUT("/api/v1/town/:town/automod", makeHandle(townPutAutoMod))
	//	automod users
	rootHandler.POST("/api/v1/town/:town/automod/user", makeHandle(townPostAutoModUser))
	rootHandler.DELETE("/api/v1/town/:town/automod/user", makeHandle(townDeleteAutoModUser))
	//	automod regexp
	rootHandler.POST("/api/v1/town/:town/automod/regexp", makeHandle(townPostAutoModRegexp))
	rootHandler.DELETE("/api/v1/town/:town/automod/regexp", makeHandle(townDeleteAutoModRegexp))

	//	Posts
	rootHandler.GET("/api/v1/posts/", makeHandle(postsGet))

	//user
	rootHandler.GET("/api/v1/user/:user/", makeHandle(userGet))
	rootHandler.POST("/api/v1/user/", makeHandle(userPost))
	rootHandler.PUT("/api/v1/user/:user/", makeHandle(userPut))
	rootHandler.PUT("/api/v1/user/:user/image/", makeHandle(userPutImage))
	rootHandler.GET("/api/v1/user/:user/image/", makeHandle(userGetImage))
	//	user notifications
	rootHandler.GET("/api/v1/user/:user/notifications/", makeHandle(userGetNotifications))
	rootHandler.PUT("/api/v1/user/:user/notifications/", makeHandle(userPutNotifications))
	rootHandler.POST("/api/v1/user/:user/notifications/", makeHandle(userPostNotifications))
	// 	user towns
	rootHandler.GET("/api/v1/user/:user/town/", makeHandle(userGetTowns))
	rootHandler.POST("/api/v1/user/:user/town/:town", makeHandle(userPostTown))     //join town
	rootHandler.DELETE("/api/v1/user/:user/town/:town", makeHandle(userDeleteTown)) //leave town
	// 	user posts
	rootHandler.GET("/api/v1/user/:user/post/", makeHandle(userGetPosts))
	//	user saved posts
	rootHandler.GET("/api/v1/user/:user/post/saved/", makeHandle(userGetSavedPosts))
	rootHandler.POST("/api/v1/user/:user/post/saved/:post", makeHandle(userPostSavedPost))
	rootHandler.DELETE("/api/v1/user/:user/post/saved/:post", makeHandle(userDeleteSavedPost))
	//	user comments
	rootHandler.GET("/api/v1/user/:user/comment/", makeHandle(userGetComments))
	//	user email confirmation
	rootHandler.PUT("/api/v1/user/:user/confirmemail", makeHandle(userConfirmEmail))
	//	user match search
	rootHandler.GET("/api/v1/user/", makeHandle(userGetMatch))

	//email test
	rootHandler.GET("/api/v1/email/", makeHandle(emailGet))

	//sessions and 3rd party
	rootHandler.POST("/api/v1/session/", makeHandle(sessionPost))
	rootHandler.DELETE("/api/v1/session/", makeHandle(sessionDelete))

	rootHandler.GET("/api/v1/session/3rdparty/", makeHandle(thirdPartyGet))
	rootHandler.POST("/api/v1/session/3rdparty/", makeHandle(thirdPartyPost))

	rootHandler.GET("/api/v1/session/facebook/", makeHandle(facebookGet))
	rootHandler.POST("/api/v1/session/facebook/", makeHandle(facebookPost))
	rootHandler.DELETE("/api/v1/session/facebook/", makeHandle(facebookDelete))

	rootHandler.GET("/api/v1/session/google/", makeHandle(googleGet))
	rootHandler.POST("/api/v1/session/google/", makeHandle(googlePost))
	rootHandler.DELETE("/api/v1/session/google/", makeHandle(googleDelete))

	rootHandler.GET("/api/v1/session/twitter/", makeHandle(twitterGet))
	rootHandler.POST("/api/v1/session/twitter/", makeHandle(twitterPost))
	rootHandler.DELETE("/api/v1/session/twitter/", makeHandle(twitterDelete))

	//images
	rootHandler.GET("/api/v1/image/:image", makeNoZipHandle(imageGet))
	rootHandler.POST("/api/v1/image/", makeHandle(imagePost))

	//posts
	rootHandler.GET("/api/v1/post/:post", makeHandle(postGet))
	rootHandler.POST("/api/v1/post/", makeHandle(postPost))
	rootHandler.PUT("/api/v1/post/:post", makeHandle(postPut))

	//post comments
	rootHandler.GET("/api/v1/post/:post/comment/", makeHandle(commentGet))
	rootHandler.GET("/api/v1/post/:post/comment/:comment", makeHandle(commentGet))
	rootHandler.POST("/api/v1/post/:post/comment/", makeHandle(commentPost))
	rootHandler.POST("/api/v1/post/:post/comment/:comment", makeHandle(commentPost))
	//rootHandler.PUT("/api/v1/post/:post/comments/:comment", makeHandle(commentPut)) // updating comments?

	//search
	rootHandler.GET("/api/v1/search", makeHandle(postSearchGet))

	if devMode {
		//handy short b64 uuid to long uuid for development
		rootHandler.GET("/api/v1/uuid/:uuid", makeHandle(uuidGet))
	}

	rootHandler.POST("/api/v1/forgotpassword", makeHandle(forgotPassword))
	// reset password requires a special handler so as to not cause CSRF issues with password resets
	rootHandler.PUT("/api/v1/forgotpassword/:token", resetPassword)

	//contact
	rootHandler.POST("/api/v1/contact", makeHandle(contactMessage))

	return rootHandler
}
