---
slug: using-cla-assistant-with-ferretdb
title: "Using CLA Assistant with FerretDB"
author: Alexey Palazhchenko
image: ../static/img/blog/cla3.jpg
date: 2022-05-16
---

![CLA Assistant](../static/img/blog/cla3.jpg)

<!--truncate-->

Like many other open-source projects, FerretDB requires all contributors to sign [our Contributor License Agreement (CLA)](https://gist.github.com/ferretdb-bot/554e6a30bfcc1d954f3853b4aad95281) to protect them from liability.
(Please note that our CLA does include a transfer of copyright and we don’t use it to relicense FerretDB; but that all is a topic of the future blog post.)

Signatures can be collected manually or with some custom scripts, but there is also a popular fully automated solution that lowers the barrier for contributors – [CLA Assistant](https://cla-assistant.io).
That software is [open-source](https://github.com/cla-assistant/cla-assistant) and uses any MongoDB-compatible database.

Recently, we released FerretDB 0.2 which implements enough functionality for CLA Assistant to work with our database without changes.
Although FerretDB is [not production-ready yet](https://github.com/FerretDB/FerretDB/#scope-and-current-state), we are big fans of dogfooding, so we already run our own instance at [cla.ferretdb.io](https://cla.ferretdb.io) and use it in FerretDB development.
In this blog post, we describe how you can host your installation using only open-source software.

Let’s start with FerretDB and PostgreSQL.
We will use Docker Compose to run everything in Docker containers.
Put the following into the docker-compose.yml file:

```js
services:
  postgres:
    image: postgres:14.2
    environment:
      POSTGRES_DB: ferretdb
      POSTGRES_HOST_AUTH_METHOD: trust
    volumes:
      - ./data/postgres:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:0.2.0
    restart: on-failure
    command: >
      -listen-addr=:27017
      -postgresql-url=postgres://postgres@postgres:5432/ferretdb

```

The first service starts PostgreSQL and creates “ferretdb” database, with data stored on the host system in “*./data/postgres*” directory.
That ensures that data is not lost when you recreate this Compose project and makes the simplest way to do backups (by just copying this directory) possible.
Of course, without more advanced backup solutions and with authentication disabled, that’s not a fully production-ready deployment, but good enough for an example.

The second service starts FerretDB which would connect to this PostgreSQL instance and listen on the standard MongoDB port.
FerretDB starts very fast and exits if it can’t connect to PostgreSQL; “*restart: on-failure*” ensures that it is restarted in that case.

Now we need to start CLA Assistant itself.
They do not provide a prebuilt Docker image, but it is easy to build ourselves.
Run the following commands to do that:

```js
git clone https://github.com/cla-assistant/cla-assistant.git
cd cla-assistant
git checkout v2.13.0
docker build --tag cla-assistant-local .
```

That will produce a Docker image with tag “*cla-assistant-local:latest*” that you could see in the “*docker ls*” output.

Next, we will need to register an OAuth App [there](https://github.com/settings/developers) that will be used by CLA Assistant to receive webhooks from pull requests:

![Register an Oauth App](../static/img/blog/cla1.jpg)

App’s Authorization callback URL should be *`https://<domain>/auth/github/callback`*

We also should register a [machine user account (a.k.a. bot)](https://docs.github.com/en/get-started/learning-about-github/types-of-github-accounts#personal-accounts) on GitHub and get a personal access token [there](https://github.com/settings/tokens) that will be used to call GitHub API on behalf of not authenticated users:

![Get personal token access](../static/img/blog/cla2.jpg)

The only required scope is “*public_repo*”.

Now, let’s add CLA Assistant to our Docker Compose configuration:

```js
services:
  # postgres and ferretdb above

  cla-assistant:
    image: cla-assistant-local:latest
    restart: on-failure
    environment:
      HOST: <domain>
      PORT: 5000
      PROTOCOL: https
      MONGODB: mongodb://ferretdb:27017/cla_assistant
      GITHUB_CLIENT: <OAuth App's Client ID>
      GITHUB_SECRET: <OAuth App's Client secret>
      GITHUB_ADMIN_USERS: <bot's account name>
      GITHUB_TOKEN: <bot's personal access token>

```

Finally, we need a web server that would handle HTTPS for us.
For that, we will use [Caddy](https://caddyserver.com):

```js
services:
  # postgres, ferretdb, and cla-assistant above

    image: caddy:2.4.6
    ports:
      - 80:80
      - 443:443
    volumes:
      - ./data/caddy/data:/data
      - ./data/caddy/config:/config
      - ./Caddyfile:/etc/caddy/Caddyfile:ro

```

Caddy will listen on both HTTP and HTTPS ports, and retrieve the TLS certificate from Let’s Encrypt that will be stored in “./data/caddy” on the host.
For that, we need to create a file called “Caddyfile” on the host next to docker-compose.yml with the following content:

```js
  <domain> {
    reverse_proxy cla-assistant:5000
    tls <your email address>
  }

```

Email is used by Let’s Encrypt to contact you if [something goes wrong](https://letsencrypt.org/docs/expiration-emails/).

That’s all with the configuration!
Now we can start our containers with *docker-compose up --detach*, start following logs with *docker-compose logs -f*, and open our domain in the browser to login with GitHub and configure our first CLA.

Hopefully, both CLA Assistant and FerretDB will work great for you; but if you encounter any problems, or just want to give us feedback about FerretDB, feel free to [join our community Slack or any other community place](https://github.com/FerretDB/FerretDB/#community) – we will be happy to help!
