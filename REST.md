REST API Documentation
===========================

Contents
--------------

* [API Conventions](#api)
* [Authentication](#auth)
* [Town](#town)
* [Post](#post)
* [User](#user)
* [Session](#session)


<a name="api"></a>API Conventions
---------------------------------
Responses will be in the JSEND (http://labs.omniti.com/labs/jsend) format with appropriate
http status codes.

```
{
	status: <"success"|"fail"|"error">,
	data: {
		<content>	
	}
}
```

* success - Submission succeeded - usually 200 or 201
* fail - There was a problem with the data submitted - client error - usually (400)
* error - There was a problem on the server - server error - usually (500)

If status is error look for `data.message`.  It will always be there in error / failure returns. 
For multiple failures, there will be a failure item with a message for each that has failed.
```
{
	status: "fail",
	data: [
		{name: "item1"},
		{name: "item2"},
		{name: "item4"},
		{name: "item6"}
	],
	failures: [
		{message: "item 3 failed", data: {name: "item3"}},
		{message: "item 5 failed due to no reason", data: {name: "item5"}}
	]
}

```

The API is versioned.  The version number in the url paths will only change if there is a backwards incompatible change to that specific area.  Meaning if a field is removed, or the behavior has changed for a specifc path, then the path will become `/v2/<path>/` and all other API paths will remain v1 until a backwards incompatible change is made to them.  Previous versions of an API path should still retain their same response and behavior even if a new version of the path exists.

<a name="auth"></a>Authentication
---------------------------------
There are only two forms of authentication, via a session cookie, or with an API Token.




<a name="town"></a>Town
------------------------------

### Get


#### /api/v1/town
Returns all towns


#### /api/v1/town/<town-id>
Returns a specific town



<a name="user"></a>User
---------------------------------
### GET

#### /api/v1/user/<key>
Get a user via their key
If no key is specified, it'll retrieve the currently logged in user.

### POST

#### /api/v1/user/
Create a new user.  

* Must specify an email + 
* password, or twitter, facebook, or google id

```
{
	email: <required>,
	password: <optional>,
	facebookID: <optional>,
	twitterID: <optional>,
	googleID: <optional>,	
}

```

### PUT

#### /api/v1/user/<key>
Update a user 

```
{
	name: <optional>,
	password: <optiona>,
}
```

#### /api/v1/user/<key>/town
Get a user's towns


#### /api/v1/user/<key>/session
Get a user's sessions

#### /api/v1/user/<key>/post
Get a user's posts

<a name="session"></a>Session
---------------------------------
### GET

### POST

#### /api/v1/session/
Post a new session.  I.E. Login

{
	email: <required>,
	password: <optional>,
	facebookID: <optional>,
	twitterID: <optional>,
	googleID: <optional>,	
	rememberMe: <optional, defaults to false>,
}


### DELETE

#### /api/v1/session/
Log out of the current session

