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

[Refer to the Coroot documentation](https://coroot.com/docs) to set up Coroot and all its components on Docker or via Helm.

In this guide, we will set up Coroot to monitor our FerretDB instance.
If you are yet to setup FerretDB, see the [FerretDB Docker installation guide](https://docs.ferretdb.io/quickstart-guide/docker/).

Deploy Coroot with the following command:

```sh
curl -fsS https://raw.githubusercontent.com/coroot/coroot/main/deploy/docker-compose.yaml | docker compose -f - up -d
```

Since Coroot is deployed locally, you can access it at http://localhost:8080/.

Depending on your setup, you may need to modify the Docker compose `yaml` file and configure Prometheus to pick up FerretDB metrics.

To update the Prometheus container's configuration file, `exec` into the container:

```sh
docker exec -it <prometheus_container_name> /bin/sh
```

Then navigate to the file to edit:

```sh
cd /etc/prometheus
vi prometheus.yml
```

Add the following to the Prometheus configuration file:

```text
scrape_configs:
  - job_name: 'ferretdb'
    metrics_path: '/debug/metrics'
    static_configs:
      - targets: ['<ferretdb-container>:8088']
```

Ensure to replace `<ferretdb-container>` with the FerretDB container name or IP address and then restart the Prometheus container so it can pick up the new configuration.

```sh
docker restart <prometheus_container_name>
```

Comfirm that the FerretDB metrics are being collected by Prometheus by navigating to the Prometheus targets at http://localhost:9090/targets.

![Prometheus targets](/img/blog/ferretdb-coroot/prometheus-targets.png)

Once the setup is complete, Coroot will start collecting metrics from FerretDB.

## Monitor FerretDB performance with Coroot

The Coroot dashboard provides the full details on all components.

Ensure Prometheus integration is correctly set up to collect metrics from FerretDB.

![Coroot prometheus integration](/img/blog/ferretdb-coroot/prometheus-integration.png)

From the Coroot dashboard, you can view the FerretDB metrics through Prometheus.

![CPU dashboard 1](/img/blog/ferretdb-coroot/cpu-metrics-1.png)

![CPU dashboard 1](/img/blog/ferretdb-coroot/cpu-metrics-2.png)

In the images, the FerretDB instance indicates a peak Requests Per Second (RPS) of 0.07 with a consistent 2ms latency.

![memory usage](/img/blog/ferretdb-coroot/memory-metrics.png)

Looking at the memory usage metrics.
you can see that the system is effectively managing memory resources without any significant performance degradation.
They show a gradual increase in usage, peaking at around 20MB without any out-of-memory issues or significant leaks.

We can also monitor the Postgres database to know how it interacts and handles all requests from FerretDB.

![Postgres dashboard](/img/blog/ferretdb-coroot/postgres-cpu-1.png)

![Postgres dashboard](/img/blog/ferretdb-coroot/postgres-cpu-2.png)

The Postgres container indicates a sharp CPU usage spike around 18:00, mirrored by a similar increase in CPU delay.
Despite the spike, throttled time remains negligible.
Node CPU usage also show moderate fluctuations but stays within acceptable limits.
So far, we can see that the system can handle peak loads without significant performance loss.

It also demonstrates stable I/O performance with low latency around 0.5ms and consistent I/O utilization peaking at 4%.
IOPS remains steady at approximately 75-100 and the bandwidth usage stays around 2-3MB/s.

## Analyze and optimize FerretDB performance with Coroot

Coroot provides a comprehensive view of your entire system, including FerretDB and all its components.
You can monitor and analyze performance metrics in real-time, identify bottlenecks, and optimize your system for better performance.
It also includes predefined inspections of your system and its components, including SLO error rate/latency and notifications.
You also get CPU, memory, storage, network, and log management metrics.

To get started with FerretDB, [see our documentation](https://docs.ferretdb.io/).
And if you want to contact the team for help or have any questions, [contact us on any of our community channels](https://docs.ferretdb.io/#community).
