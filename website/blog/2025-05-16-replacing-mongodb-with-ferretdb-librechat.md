---
slug: replacing-mongodb-with-ferretdb-librechat
title: 'Replacing MongoDB with FerretDB in LibreChat'
authors: [alex]
description: >
  A guide to replacing MongoDB with FerretDB for a fully open-source LibreChat setup.
image: /img/blog/ferretdb-librechat.jpg
tags:
  [
    tutorial,
    document databases,
    community,
    open source,
    compatible applications
  ]
---

![Replacing MongoDB with FerretDB on LibreChat](/img/blog/ferretdb-librechat.jpg)

[LibreChat](https://www.librechat.ai/) is a free, open-source application that provides a user-friendly and customizable interface for interacting with various AI providers.
It allows users to connect with providers like [OpenAI](https://openai.com/), [Azure](https://azure.microsoft.com/), [Anthropic](https://www.anthropic.com/), and more.

For LibreChat users who want to stay fully open source, FerretDB is a great drop-in replacement for MongoDB,
especially if you're looking to avoid proprietary databases or vendor lock-in.
It uses PostgreSQL with DocumentDB extension as the backend, while letting you use familiar MongoDB operations and commands.

This guide shows how to run LibreChat with FerretDB as the database, either by connecting to an existing FerretDB instance or running everything in Docker.

<!--truncate-->

## Prerequisites

Make sure you have a working LibreChat setup before proceeding.
You'll also need API keys for your preferred AI providers.
This guide uses `OPENAI_API_KEY` which is already set in the `.env` file.

## How to use FerretDB with LibreChat

This guide assumes you have already have a running instance of LibreChat.
See the [LibreChat documentation](https://www.librechat.ai/docs/quick_start/local_setup) for installation instructions.

LibreChat can connect to FerretDB in two ways:

### Option 1: Connect to an existing FerretDB instance

If you already have FerretDB running, simply replace the MongoDB URI with that of the FerretDB instance using the `MONGO_URI` environment variable.

For Docker-based setups, update the `docker-compose.override.yml` file:

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://<username>:<password>@<host>:<port>/LibreChat
```

For local development with `npm`, update the `.env` file to point to your FerretDB instance instead of MongoDB.

For example:

```text
MONGO_URI=mongodb://<username>:<password>@<host>:<port>/LibreChat
```

:::note

If you're new to FerretDB, you can find [installation instructions here](https://docs.ferretdb.io/installation/ferretdb/).

:::

### Option 2: Add FerretDB and PostgreSQL via Docker Compose

If you don't have FerretDB running, you can run it alongside LibreChat using Docker Compose.

To do that, add FerretDB and PostgreSQL with DocumentDB extension to your `docker-compose.override.yml` file.

Here's an example:

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://<username>:<password>@ferretdb:27017/LibreChat

  postgres:
    image: ghcr.io/ferretdb/postgres-documentdb:17-0.102.0-ferretdb-2.1.0
    platform: linux/amd64
    restart: on-failure
    environment:
      - POSTGRES_USER=<username>
      - POSTGRES_PASSWORD=<password>
      - POSTGRES_DB=postgres
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:2.1.0
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://<username>:<password>@postgres:5432/postgres
```

Replace `<username>` and `<password>` with your desired FerretDB credentials.

Once set up, run `docker compose up` to start the entire stack.
This will start FerretDB and PostgreSQL with DocumentDB extension, and LibreChat will connect to FerretDB using the specified connection string.

This setup allows you to run LibreChat in a fully open-source environment without vendor lock-in or license restrictions.

## Interacting with your AI providers via LibreChat

After starting the services, you can access LibreChat by navigating to `http://localhost:3080` in your web browser.
This will open up the LibreChat interface, where you can sign up and proceed to interact with your AI providers or models.

The image below shows an interaction with OpenAI through the LibreChat interface:

![LibreChat with OpenAI](/img/blog/librechat-interface.jpg)

Once you interact with LibreChat, it creates a database named `LibreChat` in FerretDB and stores all user conversations and settings there.
You can verify this by listing the collections:

```text
LibreChat> show collections
actions
agents
assistants
balances
banners
conversations
conversationtags
files
keys
messages
pluginauths
presets
projects
promptgroups
prompts
roles
sessions
sharedlinks
tokens
toolcalls
transactions
users
LibreChat>
```

You can view the conversation data in `messages` collection:

```text
LibreChat> db.messages.find().sort({ createdAt: -1 }).limit(2).pretty()
[
  {
    _id: ObjectId('6825eddb00716476ac090380'),
    messageId: '934ef840-1bb9-46ba-87e1-3df1c3606de8',
    user: '6824eb05e1a5e446598cb7e3',
    updatedAt: ISODate('2025-05-15T13:36:27.060Z'),
    expiredAt: null,
    unfinished: false,
    tokenCount: 145,
    text: 'Sure! Here are three major sights to see in Brazil:\n' +
      '\n' +
      '1. **Christ the Redeemer (Cristo Redentor)** – Located in Rio de Janeiro, this iconic statue of Jesus Christ stands atop Mount Corcovado and is one of the New Seven Wonders of the World.\n' +
      '\n' +
      '2. **Iguaçu Falls (Cataratas do Iguaçu)** – Situated on the border between Brazil and Argentina, these breathtaking waterfalls are among the largest and most impressive in the world, located within Iguaçu National Park.\n' +
      '\n' +
      '3. **Amazon Rainforest** – The vast and biodiverse Amazon Rainforest can be explored from cities like Manaus. Visitors can experience unique wildlife, indigenous cultures, and the mighty Amazon River.',
    finish_reason: 'stop',
    endpoint: 'openAI',
    sender: 'GPT-4o',
    model: 'chatgpt-4o-latest',
    isCreatedByUser: false,
    parentMessageId: 'aed06022-f0c0-42cc-a223-a985bc204cfe',
    conversationId: '847501ba-85c3-4a80-ac0c-de78bb795056',
    __v: 0,
    createdAt: ISODate('2025-05-15T13:36:27.060Z'),
    error: false,
    _meiliIndex: true
  },
  {
    _id: ObjectId('6825edd8c3efbe53320a2cf0'),
    messageId: 'aed06022-f0c0-42cc-a223-a985bc204cfe',
    user: '6824eb05e1a5e446598cb7e3',
    updatedAt: ISODate('2025-05-15T13:36:27.268Z'),
    expiredAt: null,
    unfinished: false,
    endpoint: 'openAI',
    tokenCount: 18,
    isCreatedByUser: true,
    text: 'Can you list three major sights to see in Brazil?',
    sender: 'User',
    conversationId: '847501ba-85c3-4a80-ac0c-de78bb795056',
    parentMessageId: '73794ba9-d0f8-4b1c-8eea-5daea0abb69f',
    __v: 0,
    createdAt: ISODate('2025-05-15T13:36:24.580Z'),
    model: null,
    error: false,
    _meiliIndex: true
  }
]
```

With this, we can confirm LibreChat has persisted the conversation history in FerretDB's `messages` collection.
You'll see both the user's question and the assistant's responses, along with metadata like model, sender, timestamps, and token counts.

That's it!
You've successfully replaced MongoDB with FerretDB in LibreChat.

## Further resources

By swapping MongoDB for FerretDB, you can run LibreChat in a completely open-source setup – without vendor lock-in or the license restrictions that come with SSPL.

To learn more about FerretDB, check out the following resources:

- [Setup authentication for FerretDB](https://docs.ferretdb.io/security/auth/)
- [Troubleshooting FerretDB](https://docs.ferretdb.io/troubleshooting/)

Need help?
Feel free to reach out to us on any of [our community channels](https://docs.ferretdb.io/#community).
