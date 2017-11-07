Townsourced
===========

A locally moderated community bulletin board, built to help small communities discover events, share, shop and sell locally.

# Background

My goal with Townsourced was first and foremost to build something useful for local, small communities.  I felt local
communities were very under-represented online, and the networks available catered either to large communities or niche
communities of shared interest.  My target audience was small, local communities: small towns, college campuses, 
neighborhoods, churches, schools.  I wanted to build something that handled the overlap between local communities better
than existing solutions.

It started as a small side project, and grew from there.  I showed it to a few angel investors, and demoed it to the
Minneapolis Tech community at [Minnedemo](https://minnestar.org/minnedemo/), but it never went any further than that.

I've finally gotten around to open sourcing it, and I believe there are a few interesting things in the source code that
I hope will, at a minimum, be useful as an example.

More details can be found at [tech.townsourced.com](http://tech.townsourced.com) where I've dived into the deeper details
of **Townsourced**.

You can access a live version of this code running at https://www.townsourced.com.

If you have any questions, or need help running your own instance of townsourced for your own community, feel free to
ping me at tim@townsourced.com.

# Quick Start


# Building

# Deploying


# Overview

The code is split into 3 layers:

1. **web** - Routes, web server handling, json parsing, cookies, etc.
2. **app** - Application level code.  This layer will not have any references to data layer queries or web layer structures like cookies, http requests, or json parsing.  The application layer gets passed Types and returns Types.
3. **data** - Data and the handling of servers responsible for the data.  Mostly this will involve rethinkdb, but it will also contain calls for caching in memcached, and handle things like key distribution and server selection for cache keys.
