// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

/*
Package private is for private items like passwords, API keys, and anything else that shouldn't
be released to a public repository *when* townsourced is opensourced.  The idea being that the entire
code repository can be dropped into github or wherever, and only this package should be removed.

The intention is not for this package to contain "proprietary algoritms" or some other BS.  I'd like for everything
that is townsourced to be opensourced as well as to build townsourced using opensource libraries.

That being said, I can't predict what will be needed in the future, and I would much prefer to be able to easily
open source townsourced even if that means moving access to proprietary libraries into this package.

Essentially, anything that shouldn't be put in a public repository should come from this package.
*/
package private

/* Facebook API keys */
const (
	FacebookAppID    = "redacted"
	FacebookDevAppID = "redacted"

	FacebookProdClientSecret = "redacted"
	FacebookDevClientSecret  = "redacted"
)

/* Twitter API Keys */
const (
	TwitterAPIKey    = "redacted"
	TwitterAPISecret = "redacted"
)

/* Google API Keys */
const (
	GoogleClientID     = "redacted"
	GoogleClientSecret = "redacted"
	GoogleMapsAPIKey   = "redacted"
)

/* SendGrid API keys */
const (
	SendGridAPIKey = "redacted"
)

/* IP2Location username and password */
const (
	IP2LocationEmail    = "redacted"
	IP2LocationPassword = "redacted"
)