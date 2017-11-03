Townsourced
===========

Townsourced is a bulletin board for local communities.

# Background

My goal with Townsourced was first and foremost to build something useful for local, small communities.  I felt local
communities were very under-represented online, and the networks available catered either to large communities or niche
communities of shared interest.  My target audience was small, local communities: small towns, college campuses, 
neighborhoods, churches, schools.  

Existing

More details can be found at tech.townsourced.com

# Quick Start


# Building

# Deploying


# Overview

The code is split into 3 layers:

1. web - Routes, web server handling, json parsing, cookies, etc.
2. app - Application level code.  This layer will not have any references to data layer queries or web layer structures like cookies, http requests, or json parsing.  The application layer gets passed Types and returns Types.
3. data - Data and the handling of servers responsible for the data.  Mostly this will involve rethinkdb, but it will also contain calls for caching in memcached, and handle things like key distribution and server selection for cache keys.
