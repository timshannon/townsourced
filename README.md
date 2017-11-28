# Townsourced
#### *A locally moderated community bulletin board, built to help small communities discover events, share, shop and sell locally.*

# Background

My goal with Townsourced was first and foremost to build something useful for local, small communities.  I felt local
communities were very under-represented online, and the networks available catered either to large communities or niche
communities of shared interest.  My target audience was small, local communities: small towns, college campuses, 
neighborhoods, churches, schools.  I wanted to build something that handled the overlap between local communities better
than existing solutions.

It started as a small side project, and grew from there.  I showed it to a few angel investors, and demoed it to the
Minneapolis Tech community at [Minnedemo](https://minnestar.org/minnedemo/), but it never went any further than that.

I firmly believe that open-source is important, and I want to contribute back in any way I can. There are a few 
interesting things in Townsourced's code that I hope will, at a minimum, be useful as an example.

More details can be found at [tech.townsourced.com](http://tech.townsourced.com) where I've dived into the deeper details
of **Townsourced**.

You can access a live version of this code running at https://www.townsourced.com.

If you have any questions, or need help running your own instance of townsourced for your own community, or for fun feel
free to send me a message at tim@townsourced.com.

# Quick Start

The quickest way to run Townsourced is with Docker.  Install [Docker](https://www.docker.com/) and 
[Docker-Compose](https://docs.docker.com/compose/install/), then run `docker-compose up`.

This will start a containerized instance of townsourced listening on port `8080`, running with RethinkDB, Elasticsearch
and Memcached.

# Building

Townsourced is written in go, and all of it's dependencies are vendored.  You can build townsourced by simply running
`go get github.com/timshannon/townsourced`.

To build the web / client portion you'll need to install [gobble](https://github.com/gobblejs/gobble)

`npm install -g gobble-cli`

Then run `gobble build static -f` in the `web` folder.

# Deploying

The easiest way to deploy townsourced is to use the Docker.  The `docker-compose.yaml` file can be used to run the
entire stack, or give you an insight into what services are needed.  Simply run `docker-compose up` and you'll have
a running instance of townsourced at http://localhost:8080.

To deploy Townsourced, you'll need:
* [RethinkDB](https://rethinkdb.com/)
* [Elasticsearch](https://www.elastic.co/products/elasticsearch) (2.4.6)
* [Memcached](https://memcached.org/)

Townsourced will look for a `settings.json` file either in it's running directory, or in /etc/townsourced/ on linux. In
that settings.json file will be the connection parameters that Townsourced will use to connect to the various services.
Here is an example of what that looks like:
```json
{
    "app": {
        "httpClientTimeout": "30s",
        "taskPollTime": "1m",
        "taskQueueSize": 100
    },
    "data": {
        "cache": {
            "addresses": [
                "127.0.0.1:11211"
            ]
        },
        "db": {
            "address": "127.0.0.1:28015",
            "database": "townsourced",
            "timeout": "60s"
        },
        "search": {
            "addresses": [
                "http://127.0.0.1:9200"
            ],
            "index": {
                "name": "townsourced",
                "replicas": 1,
                "shards": 5
            },
            "maxRetries": 0
        }
    },
    "web": {
        "address": "https://www.townsourced.com",
        "certFile": "",
        "keyFile": "",
        "maxHeaderBytes": 0,
        "maxUploadMemoryMB": 10,
        "minTLSVersion": 769,
        "readTimeout": "60s",
        "writeTimeout": "60s"
    }
}
```

Finally, you'll need a `web/static` folder (built from gobble) in the running directory of townsourced.


# Overview

The code is split into 3 layers:

1. **web** - Routes, web server handling, json parsing, cookies, etc.
2. **app** - Application level code.  This layer will not have any references to data layer queries or web layer structures like cookies, http requests, or json parsing.  The application layer gets passed Types and returns Types.
3. **data** - Data and the handling of servers responsible for the data.  Mostly this will involve rethinkdb, but it will also contain calls for caching in memcached, and handle things like key distribution and server selection for cache keys.


More detail [here](http://tech.townsourced.com/post/anatomy-of-a-go-web-app/).

# API Keys

In the `data/private/private.go` file is where you can enter your own API keys for Google, Facebook, Twitter, and SendGrid.
Without these keys, things like google maps integration, and sending email will not work.  See the `private.go` file for
more details.
