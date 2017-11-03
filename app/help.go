// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import "github.com/timshannon/townsourced/data"

// Help is a help document entry
type Help struct {
	Key      data.Key `json:"key"`
	Title    string   `json:"title"`
	Document string   `json:"document"`
}

// HelpGet Retrieves a help document
func HelpGet(key data.Key) (*Help, error) {
	return &Help{
		Key:      key,
		Title:    "FAQ",
		Document: helpPlaceholder,
	}, nil
}

// this is a placeholder until the full help system is put into place
const helpPlaceholder = `
# Customize your Profile
Welcome to Townsourced! Setting up your profile is an important step to being an active member of of your Townsourced Towns. 

## Why is it important to set up my profile?
* Change your Name
	* Townsourced allows you to show only the name you want to be seen. 
	* If you prefer a nickname or to have your last name abbreviated to just an initial, you can do so at the top of your profile image by clicking the pencil icon. 
* Verify your email address
	* Verifying email address allows Townsourced to be sure that when notifications are sent to your email, they are actually going to you.
	* You can verify your email address by clicking “verify email address” underneath your profile picture.
* Profile image
	* You do not have to add a profile image, it is recommended though. 
	* Townsourced recommends buyers and sellers to look at users profiles before meeting up with other users as a safety precautions. Read more about that under **Safety Tips**.
 
            
# Tips to Help Your Post Get Noticed
##  Add a picture!
A picture is worth 1,000 words, so make use of Townsourced’s add a photo feature! Clear photos with uncluttered backgrounds will best represent your post.
### You can add pictures two ways.
1. Click the gray “Select an Image” box and choose a picture from your desktop or phone.
1. You can also drag and drop a picture from your computer. 


## Add \#Hashtags so people can easily discover your post!
\#Hashtags are a great way to help people find your post. The more \#Hashtags you add that are relevant to your post, the easier your post will be found. The most common \#hashtags that are used in Townsourced will pop up as you type. 

### You can insert \#hashtags in two ways
1. Type “ # “ into the text box
	* Start typing in your hashtag 
	* You can also choose from the popup selection of the most used hashtags.
1. Insert a common hashtag image
	1. Click the “smiley face emoji” tab
	1. Select “Tags” icon.
	1. Choose a Tag. 
	1. If you are not sure what the Hashtags mean, hover over the Hashtag and a tool tip will inform you.
    
## Try a different Layout!
Not every post works best with the “Standard” layout. Townsourced allows you to customize your post so you can best represent what you are posting about.

After you write your post click on the “Preview” tab next to the “write” tab in the upper left corner of the text box. When you switch to the “Preview” tab you will find four layout options to choose from. You can easily view each option by clicking on a layout button.
* **Standard**
	* Standard Layout puts the text on top and the images at the bottom. This is the format that will be selected for your post if you opt not to customize.
* **Article**
	* Think of an Article like reading a newspaper or magazine. If you have a lot of text to get across but still want to feature one image, Article is the best fit for your post. 
* **Gallery**
	* Gallery Layout highlights the photos of your posts and lets people easily toggle through your photos. This is great for photos that show an item from different angles or also showing a wide range of items that will be sold at a Garage Sale.
* **Poster** 
	* Tryout the Poster layout if you have one great picture that tells people what you want them to know. Think of the Poster Layout like a billboard or a flyer you would stick to a bulletin board.     


# Townsourced Safety Tips

Safety is very important to Townsourced. That is why, with Townsourced, you never have to share a personal email account or phone number and your personal details are private. 

## Meet Up Safety Tips
Before meeting up with a new person, Townsourced recommends all their users to read and follow these Safety Tips.
* Check the profile of the user you planning on meeting.
	* Do they have a personal profile photo?
		* When meeting up with someone new it is good to know who you will be looking for and what to expect.
	* Note how long they have been a member. 
		* The longer a user has been a member, the more likely they will be a legitimate user.
	* Post History
		* Check a person’s post history so you know if a sale of this kind is typical or atypical for the person. Have they sold things before? What kinds of things are they posting to Townsourced?
	* Comment History    
		* What someone comments on the internet says a lot about a person. Are the user’s comments overall constructive and positive? Are they adding value to the community? Are they respectful?  These interactions can give you a good idea of what type of person you will be dealing with.
* Meeting Place
	* Check your “Towns” profile page to see if there are any recommended meet up places in your area. 
		* Meet in a well-lit, public place
		* Meet up at places with security cameras whenever possible
		* If possible, bring a friend
		* Always tell someone your plans
			* Where you will meet 
			* What you will be exchanging
			* Who you expect to meet
			* When the meeting should be done
			* Let them know when the meetup is complete
		* If possible, bring a cell phone with you

            
# Categories    
To better help filter and find posts, Townsourced has created broad categories.
* **Notices**
	* Examples of Notices
		* Town's Garbage Pickup day will be a day late due to weather
		* Lost and Found
		* General Notices that Moderators release to the “Town”
		* Town Pool is opening for the summer!
    
* **Buy & Sell**
	* Examples of Buy & Sell
		* Garage Sales
		* Piano for Sale
		* Looking to buy a used snowblower
* **Events**
	* Examples of Events
		* Upcoming community Parade
		* Local Team Fundraiser event
		* Music in the Park
		* Local Band Gigs
* **Jobs**
	* Examples of Jobs
		* Looking to hire a dog sitter
		* House Painter for Hire
		* Teenager looking for babysitter job
* **Volunteer**
	* Examples of Volunteer
		* Volunteer Opportunity for Park Clean Up
		* Looking for Volunteers to help at Food Bank
		* I am a Boy Scout looking to earn my Volunteer Badge
		* High School student looking to give back to community while beefing up College Resume
		* Ride Sharing
* **Housing**
	* Examples of Housing
		* Basement apartment for rent
		* Looking for 3 bedroom home in desirable school district
		* Looking for a roommate to help with housing costs.


`
