// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"path"
	"path/filepath"
)

// initializes all templated messages like notifications and emails
func initMessages(runningDir string) {
	emailTemplatePath := filepath.Join(runningDir, "/web/static/email/")

	//******************   comments   ***************************

	addMessageType("msgCommentUserMention", message{
		subject: `You have been mentioned in a comment on the post "{{.Post.Title}}"`,
		body: `You have been mentioned in a comment on the post **{{.Post.Title}}**. You can view the comment [here](/post/{{FromUUID .Post.Key}}/comment/{{FromUUID .Comment.Key}}).


*You can disable this type of notification from the settings tab in your user profile.*

`})

	addMessageType("msgCommentPost", message{
		subject: `A new comment has been added to your post "{{.Post.Title}}"`,
		body: `
A new comment has been added by [{{.Comment.Username}}](/user/{{.Comment.Username}}) to your post **{{.Post.Title}}**. 

> {{.Comment.Comment}}

You can view the comment [here](/post/{{FromUUID .Post.Key}}/comment/{{FromUUID .Comment.Key}}#comments).

`})

	addMessageType("msgCommentReply", message{
		subject: `Someone has replied to your comment on post "{{.Post.Title}}"`,
		body: `
[{{.Comment.Username}}](/user/{{.Comment.Username}}) has replied to your comment on post **{{.Post.Title}}**. 

> {{.Comment.Comment}}

You can view the comment [here](/post/{{FromUUID .Post.Key}}/comment/{{FromUUID .Comment.Key}}#comments).

`})

	addMessageType("emailCommentMention", message{
		subject:  `You've been mentioned in a comment on the post "{{.Post.Title}}"`,
		bodyPath: path.Join(emailTemplatePath, "commentMention.template.html"),
	})
	addMessageType("emailCommentPost", message{
		subject:  `A new comment has been added to your post "{{.Post.Title}}"`,
		bodyPath: path.Join(emailTemplatePath, "commentPost.template.html"),
	})
	addMessageType("emailCommentReply", message{
		subject:  `Someone has replied to your comment on post "{{.Post.Title}}"`,
		bodyPath: path.Join(emailTemplatePath, "commentReply.template.html"),
	})

	//******************   posts   ***************************

	addMessageType("msgPostModerated", message{
		subject: `Your post "{{.Post.Title}}" has been moderated in {{.Town.Name}}`,
		body: `
I have moderated your post {{.Post.Title}} in the town {{.Town.Name}} for the following reason:

> {{.Reason}}

You can view the post and comments [here](/post/{{FromUUID .Post.Key}}).
`})

	addMessageType("msgPostReport", message{
		subject: `I am reporting the post "{{.Title}}"`,
		body: `
I am reporting the post **{{.Title}}** for the following reason:

> {{.Reason}}

This post has been reported **{{len .Reported}} {{if len .Reported | eq 1 }} time{{else}} times{{end}}**. You can view the post and comments [here](/post/{{FromUUID .Key}}).

*This message was generated from the report button on the above post, and was sent to you because you are the moderator of one or more of the towns, the post was submitted to.*
`})

	addMessageType("msgPostUserMention", message{
		subject: `You have been mentioned in the post "{{.Title}}"`,
		body: `
You have been mentioned in the post **{{.Title}}**. 
You can view the post and comments [here](/post/{{FromUUID .Key}}).


*You can disable this type of notification from the settings tab in your user profile.*

`})

	addMessageType("emailPostMention", message{
		subject:  `You've been mentioned in the post "{{.Post.Title}}"`,
		bodyPath: path.Join(emailTemplatePath, "postMention.template.html"),
	})

	//******************   towns   ***************************
	addMessageType("msgTownModInvite", message{
		subject: `You have been invited to become a moderator of {{.Name}}`,
		body: `
I am inviting you to become a moderator of **{{.Name}}**.  To accept your invitation, go to the [town's page](/town/{{.Key}}) and review the details in the towns information panel, then click the **Accept Moderator Invite** button to become a moderator of the town.

*This message was generated from the {{.Name}} town settings page.*
`})

	addMessageType("msgTownNew", message{
		subject: `{{.Name}} has been registered`,
		body: `
Congratulations on registering a new town!  You are now the moderator of *{{.Name}}*.

As moderator your job is to maintain your town by keeping a close eye on what gets posted.  Inside of {{.Name}} you'll have access to hide a post from the public by marking it as **moderated**.  After which it will only be visible to other moderators, and the original submitter of the post.  

Types of posts you may want to mark as **moderated** include:
* SPAM, or other unsolicited advertising
* Illegal activity
* Hate speech, harassment or bullying
* Content that doesn't belong in your town (e.g. a post looking for work in North Dakota, to a town in Minnesota) 

You will also receive notifications of posts submitted to your town which been reported by users.

You can invite others to help moderate from the [town settings page](/town/{{.Key}}/settings).  Once invited their moderator rights cannot be revoked.  A moderator can only remove their own right to moderate a town.
`})

	addMessageType("msgTownPrivateInvite", message{
		subject: `You have been invited to become a member of {{.Name}}`,
		body: `
I am inviting you to become a member of **{{.Name}}**.  The town is Private. 
You can visit the [town's page](/town/{{.Key}}), to find out more, or click *Join Town* below to become a member.


[Join Town](/town/{{.Key}}?join "button")

*This message was generated from the {{.Name}} town settings page.*
`})

	addMessageType("msgTownPrivateRequestInvite", message{
		subject: `I am requesting access to the private town of {{.Name}}`,
		body: `
I am requesting to become a member of the private town **[{{.Name}}](/town/{{.Key}})**.

*This message was generated from the {{.Name}} town page, and you have been received this because you are a moderator of {{.Name}}.*

*To grant access to join this town, go to the [Privacy tab on the Town Settings page](/town/{{.Key}}/settings).*
`})

	addMessageType("msgTownPrivateAcceptInviteRequest", message{
		subject: `Your request to join the private town of {{.Name}} has been accepted`,
		body: `
Your request to join the private town **[{{.Name}}](/town/{{.Key}})** as been **accepted**, and you are now a member!

*This message was generated from the {{.Name}} town page.*
`})

	addMessageType("msgTownPrivateRejectInviteRequest", message{
		subject: `Your request to join the private town of {{.Name}} has been rejected`,
		body: `
Your request to join the private town **[{{.Name}}](/town/{{.Key}})** as been **rejected**.

*This message was generated from the {{.Name}} town page.*
`})

	addMessageType("emailTownInvite", message{
		subject:  `You have been invited to a private town on Townsourced`,
		bodyPath: path.Join(emailTemplatePath, "townInvite.template.html"),
	})

	//******************   user   ***************************
	addMessageType("emailUserConfirmEmail", message{
		subject:  `{{if .Welcome}}Welcome to Townsourced!{{else}}Townsourced Email Confirmation{{end}}`,
		bodyPath: path.Join(emailTemplatePath, "confirmEmail.template.html"),
	})
	addMessageType("emailUserForgotPassword", message{
		subject:  `Townsourced Password Reset`,
		bodyPath: path.Join(emailTemplatePath, "forgotPassword.template.html"),
	})
	addMessageType("emailUserPrivateMessage", message{
		subject:  `You've recieved a private message: {{.Subject}}`,
		bodyPath: path.Join(emailTemplatePath, "privateMsg.template.html"),
	})
	addMessageType("msgUserWelcome", message{
		subject: `Welcome to Townsourced!`,
		body: `
### Welcome to Townsourced!

We hope you find it to be a useful and important part of your community. To help get the most of out Townsourced, you should:

* **Join some towns.** They don't have to be local to you, or even towns. Townsourced is for communities of any kind, *e.g. your softball league, dorm or neighborhood*.  [Search for more towns near you](/town/)!
* **Start a new town.**  If you can't find the town your looking for, then chances are others are looking for it as well. [Start a new community](/newtown)!

Looking for more information or help?  [Check the FAQ](/help).`,
	})

	//******************   contact   ***************************
	addMessageType("emailContact", message{
		subject:  "Contact Message: {{.Subject}}",
		bodyPath: path.Join(emailTemplatePath, "contact.template.html"),
	})

}
