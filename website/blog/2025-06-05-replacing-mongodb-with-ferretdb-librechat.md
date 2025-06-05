---
slug: replacing-mongodb-with-ferretdb-librechat
title: 'Replacing MongoDB with FerretDB in LibreChat'
authors: [alex]
description: >
  A guide to replacing MongoDB with FerretDB for a fully open-source LibreChat setup that runs the Ollama phi4-mini model locally.
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

[LibreChat](https://www.librechat.ai/) is a free, open-source application that provides a user-friendly and customizable interface for interacting with various AI providers and models.

<!--truncate-->

It allows users to connect with cloud providers like [OpenAI](https://openai.com/), [Azure](https://azure.microsoft.com/), [Anthropic](https://www.anthropic.com/), and others as well as fully open-source tools like [Ollama](https://ollama.com/), which lets you run models like `phi4-mini` locally.

For LibreChat users who want to stay fully open source, FerretDB is a great drop-in replacement for MongoDB,
especially if you're looking to avoid proprietary databases or vendor lock-in.
It uses PostgreSQL with DocumentDB extension as the backend, while letting you use familiar MongoDB operations and commands.

This guide shows how to run LibreChat with FerretDB (instead of MongoDB) and Ollama to run the `phi4-mini` model locally – with all components being fully open source.

## Prerequisites

To follow this guide, ensure:

- [Docker](https://www.docker.com/) is installed and running on your machine.
- Ollama is installed and running on your machine.
  If you haven't installed Ollama yet, download it from the [official website](https://ollama.com/download).
  This guide uses the `phi4-mini` model, an open-source model optimized for reasoning and accuracy in text-based tasks.
  Once Ollama is running, pull the `phi4-mini` model:

  ```sh
  ollama pull phi4-mini
  ```

## How to use FerretDB with LibreChat

Start by cloning the LibreChat repository:

```sh
git clone https://github.com/danny-avila/LibreChat.git
cd LibreChat
```

You can find more instructions on how to set up LibreChat in the [LibreChat documentation](https://www.librechat.ai/docs/quick_start/local_setup).

Next, copy the `.env.example` file to `.env`:

```sh
cp .env.example .env
```

You may need to adjust the environment variables in `.env` to suit your setup.

You can run FerretDB alongside LibreChat using Docker Compose.

To do that, add FerretDB and PostgreSQL with DocumentDB extension to your `docker-compose.override.yml` file, as shown below:

```yaml
services:
  api:
    environment:
      - MONGO_URI=mongodb://<username>:<password>@ferretdb:27017/LibreChat
    depends_on:
      - ferretdb

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb-eval:2
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - POSTGRES_USER=<username>
      - POSTGRES_PASSWORD=<password>
      - POSTGRES_DB=postgres

  mongodb:
    profiles:
      - donotstart
```

Replace `<username>` and `<password>` with your desired FerretDB credentials.

In the above `docker-compose.override.yml` file, we use the FerretDB evaluation image (`ferretdb-eval:2`), which is suitable for development and testing purposes, but not recommended for production use.
It includes the PostgreSQL database with DocumentDB extension, which FerretDB uses as its backend.

We also place the `mongodb` service under the `donotstart` profile so it won't start by default.
This means you need to explicitly update the LibreChat `api` service to not depend on the `mongodb` service in your `docker-compose.yml` file – or you can remove the `mongodb` service entirely from the `docker-compose.yml` file.

If you're new to FerretDB, you can learn more about the [installation instructions here](https://docs.ferretdb.io/installation/ferretdb/).

## Using Ollama with LibreChat

LibreChat supports Ollama, allowing you to run open-source models like `phi4-mini` locally.
We assume you have Ollama installed and the `phi4-mini` model pulled.
You can use any other Ollama model as well, just replace `phi4-mini` with the desired model name.
You can also check the [LibreChat documentation for more details on setting up Ollama locally](https://www.librechat.ai/docs/configuration/librechat_yaml/ai_endpoints/ollama).

Start by copying the example configuration file for LibreChat:

```sh
cp librechat.example.yaml librechat.yaml
```

Then add the following under the `custom` section in `librechat.yaml`:

```yaml
custom:
  - name: 'Ollama'
    apiKey: 'ollama'
    baseURL: 'http://host.docker.internal:11434/v1/'
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
You will also have Ollama running with the `phi4-mini` model available.

This setup allows you to run LibreChat in a fully open-source environment without vendor lock-in or license restrictions.

## Interacting with your AI models via LibreChat

After starting the services, you can access LibreChat by navigating to `http://localhost:3080` in your web browser.
This will open up the LibreChat interface, where you can sign up and proceed to interact with the `phi4-mini` model running locally via Ollama.

The image below shows an interaction with the open-source `phi4-mini` model running locally via Ollama in LibreChat:

![LibreChat interaction using the phi4-mini model running locally via Ollama](/img/blog/librechat-interface.jpg)

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
    _id: ObjectId('68418b91651dcd2f71046760'),
    messageId: '2496e0e0-78dd-4618-bf03-b23fc413554b',
    user: '68418a0040fa9c1513e29565',
    updatedAt: ISODate('2025-06-05T12:20:33.528Z'),
    expiredAt: ISODate('2025-07-05T12:20:33.527Z'),
    unfinished: false,
    tokenCount: 46,
    text: '1. Christ the Redeemer - Rio de Janeiro\n' +
      '2. Sugarloaf Mountain (Pão de Açúcar) - Rio de Janeiro\n' +
      '3. Amazon Rainforest - Various locations across northern Brazil',
    finish_reason: 'stop',
    endpoint: 'ollama',
    sender: 'Ollama (phi4-mini)',
    model: 'phi4-mini',
    isCreatedByUser: false,
    parentMessageId: 'b8c08f11-913e-4528-b1ee-776a8d5c9ed1',
    conversationId: '510281d1-7cc1-4b01-b14b-f3add7302ca7',
    __v: 0,
    createdAt: ISODate('2025-06-05T12:20:33.528Z'),
    error: false,
    _meiliIndex: true
  },
  {
    _id: ObjectId('68418b8f7df82acbc800dbe0'),
    messageId: 'b8c08f11-913e-4528-b1ee-776a8d5c9ed1',
    user: '68418a0040fa9c1513e29565',
    updatedAt: ISODate('2025-06-05T12:20:33.549Z'),
    expiredAt: ISODate('2025-07-05T12:20:33.548Z'),
    unfinished: false,
    endpoint: 'ollama',
    tokenCount: 15,
    isCreatedByUser: true,
    text: 'Can you list three major sights to see in Brazil?',
    sender: 'User',
    conversationId: '510281d1-7cc1-4b01-b14b-f3add7302ca7',
    parentMessageId: '48986d58-e1e7-478f-a98b-92049bdefb67',
    __v: 0,
    createdAt: ISODate('2025-06-05T12:20:31.447Z'),
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

By swapping MongoDB for FerretDB, you can run LibreChat in a completely open-source setup, without vendor lock-in or the license restrictions that come with SSPL.
Besides, you can also run open-source models like `phi4-mini` locally using Ollama, ensuring your entire stack remains open source.

To learn more about FerretDB, check out the following resources:

- [Setup authentication for FerretDB](https://docs.ferretdb.io/security/auth/)
- [Troubleshooting FerretDB](https://docs.ferretdb.io/troubleshooting/)

Need help?
Feel free to reach out to us on any of [our community channels](https://docs.ferretdb.io/#community).
