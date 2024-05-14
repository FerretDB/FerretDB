---
slug: monitor-ferretdb-performance-using-coroot
title: 'Monitor FerretDB Performance using Coroot'
authors: [alex]
description: >
  Observability is vital in any infrastructure. Find out how Coroot can provide real-time monitoring and visibility into a FerretDB setup.
image: /img/blog/ferretdb-coroot.jpg
tags: [tutorial, community, cloud, open source]
---

![Monitor FerretDB Performance using Coroot](/img/blog/ferretdb-coroot.jpg)

Effective real-time monitoring is a critical aspect of any infrastructure.
[Coroot](https://coroot.com/) is an open source observability platform that can provide real-time monitoring and visibility into a [FerretDB](https://www.ferretdb.com/) setup.

<!--truncate-->

Effective real-time monitoring is a critical aspect of any infrastructure.
[Coroot](https://coroot.com/) is an open source observability platform that can provide real-time monitoring and visibility into a [FerretDB](https://www.ferretdb.com/) setup.
App developers can quickly troubleshoot issues without necessarily becoming experts in infrastructure operations and management.

If you're just finding out about FerretDB, it's the truly open source document database that provides MongoDB compatibility with Postgres as the backend.
With Coroot, you can get visibility into your entire system in a matter of minutes.

In this blog post, you'll learn about Coroot and how it can help you monitor and get better visibility into your FerretDB instance.

## Coroot observability for FerretDB

Setting up effective monitoring for your entire system can be resource-intensive, time-consuming, and expensive.
Imagine a scenario where your FerretDB instance is deployed as a Docker container along with other components and services.
Observability into all critical areas of the system in as little time is vital.

With Coroot, you can visualize every aspect of your entire setup.
That includes all its components, including your applications and their subsystems, databases, and, most importantly, how they all communicate.

Coroot heavily relies on eBPF, a technology that enhances observability and monitoring by eliminating the need to instrument application code manually.
Custom programs can run directly within the Linux kernel without changing kernel source code or loading additional kernel modules.
You get deeper and more extensive information on system calls, network functions, and security issues – just better observability.

## Setting up Coroot for FerretDB monitoring

Since Coroot uses eBPF, you need the right environment before setting it up.
The most recent versions of the Linux kernel (v 4.16 and above) should be compatible since they offer at least minimal eBPF support.

[Refer to the Coroot documentation](https://coroot.com/docs) to set up Coroot and all its components on Docker or via Helm.

In this guide, we will set up Coroot to monitor our FerretDB.

```yaml
volumes:
  prometheus_data: {}
  coroot_data: {}

services:
  coroot:
    image: ghcr.io/coroot/coroot
    volumes:
      - coroot_data:/data
    ports:
      - 8080:8080
    command:
      - '--bootstrap-prometheus-url=http://prometheus:9090'
      - '--bootstrap-refresh-interval=15s'
      - '--bootstrap-clickhouse-address=clickhouse:9000'
    depends_on:
      - clickhouse
      - prometheus

  node-agent:
    image: ghcr.io/coroot/coroot-node-agent
    privileged: true
    pid: 'host'
    volumes:
      - /sys/kernel/debug:/sys/kernel/debug
      - /sys/fs/cgroup:/host/sys/fs/cgroup
    command:
      - '--collector-endpoint=http://coroot:8080'
      - '--cgroupfs-root=/host/sys/fs/cgroup'

  prometheus:
    image: prom/prometheus:v2.45.4
    volumes:
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--web.enable-lifecycle'
      - '--web.enable-remote-write-receiver'
    ports:
      - '127.0.0.1:9090:9090'

  clickhouse:
    image: clickhouse/clickhouse-server:24.3
    ports:
      - '127.0.0.1:9000:9000'
    ulimits:
      nofile:
        soft: 262144
        hard: 262144

  postgres:
    image: postgres
    environment:
      - POSTGRES_USER=username
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=ferretdb
    volumes:
      - ./data:/var/lib/postgresql/data

  ferretdb:
    image: ghcr.io/ferretdb/ferretdb
    restart: on-failure
    ports:
      - 27017:27017
    environment:
      - FERRETDB_POSTGRESQL_URL=postgres://postgres:5432/ferretdb
```

The coroot-node-agent is a Prometheus exporter for gathering container metrics running on a particular node.
It gathers log-based metrics, network latency, JVM metrics, and more.

Coroot offers a distributed tracing approach that allows engineers to visualize request paths across all components.
This could help to quickly identify latency issues, errors, or bottlenecks in their setup.

Depending on your application setup, you may need to modify the Docker compose `yaml` file and configure the tracing metrics on Coroot by enabling OpenTelemetry locally.
[Check this docs to learn more](https://coroot.com/docs/coroot-community-edition/tracing/overview).

Once you apply and run the `docker compose up`, Docker starts and runs the entire application defined in the `yaml` file.

Run `docker compose ps` just to be sure it's correctly deployed.
Since Coroot was deployed locally, you can access it at http://localhost:8080/.

## Monitor FerretDB performance with Coroot

The Coroot dashboard provides the full details on all components.

At first glance, we can see a memory leak on the `ferretdb` and `postgres` databases.
That suggests that allocated memory is not being efficiently reused or deallocated, causing the total memory usage to grow progressively as the services operate.

![Coroot Dashboard](/img/blog/ferretdb-coroot/01-dashboard.png)

We can also visualize the complete system architecture via the "service map".
In the image below, you can see how our application is configured.

![Service Map](/img/blog/ferretdb-coroot/02-svc-map.png)

Through Coroot, we can monitor, visualize, and analyze essential performance metrics for our database.
It offers essential details on the database performance and other peripherals, including CPU and memory usage metrics, latency/error rates, heatmaps, and more.

![CPU metrics](/img/blog/ferretdb-coroot/03-cpu.png)

![Memory metrics](/img/blog/ferretdb-coroot/04-mem.png)

![Network metrics](/img/blog/ferretdb-coroot/05-net.png)

Coroot also provides log management features.
It gathers and displays logs from all containers on a node – error/latency rates, notification system, etc.

Below, you can see the FerretDB logs rendered in detail.

![Log metrics](/img/blog/ferretdb-coroot/06-log.png)

Using distributed tracing, Coroot provides a heat map showing operation requests, their status, durations, and details.

![Latency](/img/blog/ferretdb-coroot/07-latency.png)

The above image shows how response time for the `ferretdb` increased progressively over time.
It shows that the system takes a long time to handle queries.
That should prompt us to take additional measures to improve performance.

## Analyze and optimize FerretDB performance with Coroot

The right visualization dashboard and details to effectively and continuously monitor your FerretDB setup are just what Coroot offers.
You get predefined inspections of your system and its components, including SLO error rate/latency and notifications.
You also get CPU, memory, storage, network, and log management metrics.

To get started with FerretDB, [see our documentation](https://docs.ferretdb.io/).
And if you want to contact the team for help or have any questions, [contact us on Slack](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A).
