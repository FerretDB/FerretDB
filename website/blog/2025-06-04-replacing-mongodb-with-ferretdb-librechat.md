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
It allows users to connect with providers like [Ollama](https://ollama.com/), [OpenAI](https://openai.com/), [Azure](https://azure.microsoft.com/), [Anthropic](https://www.anthropic.com/), and others – including fully local open-source models such as phi4-mini.

For LibreChat users who want to stay fully open source, FerretDB is a great drop-in replacement for MongoDB,
especially if you're looking to avoid proprietary databases or vendor lock-in.
It uses PostgreSQL with DocumentDB extension as the backend, while letting you use familiar MongoDB operations and commands.

This guide shows how to run LibreChat with FerretDB as the database, either by connecting to an existing FerretDB instance or running everything in Docker.

<!--truncate-->

## Prerequisites

Make sure you have a working LibreChat setup before proceeding.
You'll also need API keys for your preferred AI providers.
This guide uses a local Ollama model (`phi4-mini`), which requires no API key setup.
Just make sure the model is pulled and Ollama is running.

## How to use FerretDB with LibreChat

This guide assumes you already have a running instance of LibreChat.
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
    image: ghcr.io/ferretdb/postgres-documentdb:17-0.103.0-ferretdb-2.2.0
    restart: on-failure
    environment:
      - POSTGRES_USER=<username>
      - POSTGRES_PASSWORD=<password>
      - POSTGRES_DB=postgres
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb:2.2.0
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://<username>:<password>@postgres:5432/postgres
```

Replace `<username>` and `<password>` with your desired FerretDB credentials.

## Using Ollama with LibreChat

LibreChat supports Ollama as a provider, allowing you to run open-source models like `phi4-mini` locally.
We assume you have Ollama installed and the `phi4-mini` model pulled.
You can use any other Ollama model as well, just replace `phi4-mini` with the desired model name.
You can also check the [LibreChat documentation for more details on setting up Ollama as a provider](https://www.librechat.ai/docs/configuration/librechat_yaml/ai_endpoints/ollama).

Start by copying the example configuration file for LibreChat:

```sh
cp librechat.example.yaml librechat.yaml
```

Then add the following under the `custom` section in `librechat.yaml`:

```yaml
custom:
  - name: 'Ollama'
    apiKey: 'ollama'
    baseURL: 'http://host.docker.internal:11434/v1/' # Use this if LibreChat is in Docker
    models:
      default: ['phi4-mini']
      fetch: false
    titleConvo: true
    titleModel: 'current_model'
    summarize: false
    summaryModel: 'current_model'
    forcePrompt: false
    modelDisplayLabel: 'Ollama (phi4-mini)'
```

Mount the configuration file in `docker-compose.override.yml`:

```yaml
api:
  volumes:
    - type: bind
      source: ./librechat.yaml
      target: /app/librechat.yaml
```

Then restart your containers to apply the changes:

```sh
docker compose down
docker compose up -d
```

This will start FerretDB and PostgreSQL with DocumentDB extension, and LibreChat will connect to FerretDB using the specified connection string.
You will also have Ollama running as a provider with the `phi4-mini` model available.

This setup allows you to run LibreChat in a fully open-source environment without vendor lock-in or license restrictions.

## Interacting with your AI providers via LibreChat

After starting the services, you can access LibreChat by navigating to `http://localhost:3080` in your web browser.
This will open up the LibreChat interface, where you can sign up and proceed to interact with your AI providers or models.

The image below shows an interaction with the open-source `phi4-mini` model running locally via Ollama in LibreChat:

![A LibreChat interaction with the phi4-mini model running locally via Ollama:](/img/blog/librechat-interface.jpg)

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
    _id: ObjectId('683f79f2275d40bc6a0ba610'),
    messageId: '06961e88-8192-45e5-8d9a-3ec46ca68d3b',
    user: '683ef3e3f11454fa09471eb6',
    updatedAt: ISODate('2025-06-03T22:40:50.523Z'),
    expiredAt: ISODate('2025-07-03T22:40:50.521Z'),
    unfinished: false,
    tokenCount: 43,
    text: '1. Christ the Redeemer Statue, Rio de Janeiro\n' +
      '\n' +
      '2. Sugarloaf Mountain (Pão de Açúcar), Rio de Janeiro\n' +
      '\n' +
      '3. Amazon Rainforest, Northern region',
    finish_reason: 'stop',
    endpoint: 'ollama',
    sender: 'Ollama (phi4-mini)',
    model: 'phi4-mini',
    isCreatedByUser: false,
    parentMessageId: 'dfe1c54c-4381-4e66-979b-dc8df3b958c7',
    conversationId: '5a68447f-f4e7-4156-8920-74f5833aefe7',
    __v: 0,
    createdAt: ISODate('2025-06-03T22:40:50.523Z'),
    error: false,
    _meiliIndex: true
  },
  {
    _id: ObjectId('683f79f00707f2f06a0fa8c0'),
    messageId: 'dfe1c54c-4381-4e66-979b-dc8df3b958c7',
    user: '683ef3e3f11454fa09471eb6',
    updatedAt: ISODate('2025-06-03T22:40:50.686Z'),
    expiredAt: ISODate('2025-07-03T22:40:50.681Z'),
    unfinished: false,
    endpoint: 'ollama',
    tokenCount: 15,
    isCreatedByUser: true,
    text: 'Can you list three major sights to see in Brazil?',
    sender: 'User',
    conversationId: '5a68447f-f4e7-4156-8920-74f5833aefe7',
    parentMessageId: '28c8b0d6-bbe1-4d09-8693-5b30cb691ca3',
    __v: 0,
    createdAt: ISODate('2025-06-03T22:40:48.477Z'),
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
