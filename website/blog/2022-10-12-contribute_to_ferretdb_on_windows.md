---
slug: contribute_to_ferretdb_on_windows
title: "How to start contributing to FerretDB on Windows"
author: Dmitry Eremenko
image: ../static/img/blog/image4-1024x683.jpg
date: 2022-10-12
---

![Contribute to FerretDB on Windows](../static/img/blog/image4-1024x683.jpg)

<!--truncate-->

Historically, I’ve always used Windows operating system on my laptop, and most times, it requires a bit of work to set up and configure specific development tools.
When I started working with FerretDB, I encountered several difficulties, which inspired me to write this article so that Windows users won’t have to experience the same.

For all Windows users fascinated and interested in contributing to FerretDB, this article would help resolve some of the challenges you might face.

## Prerequisites

To start contributing to FerretDB on Windows, this is a list of software we'll be using:

* Git (for example, [GitHub for Windows](https://desktop.github.com/))
* Go ([download here](https://go.dev/dl/))
* Docker for Windows ([download here](https://docs.docker.com/desktop/install/windows-install/))
* Text editor (there are plenty of them, you have to pick one! 😃)

## Git options that are extremely useful for Windows systems

* **Lines endings**: Most operating systems handle line endings differently.
To ensure effective collaboration and consistency with people using other operating systems, we need to configure git’s line endings with the following command: git config --global core.autocrlf true
* **Global .gitignore**: It would help you to avoid committing unrelated files like executables and IDE settings.
With the global .gitignore setting, you only need to set it up once, which would work for all repositories.

## Helpful Docker settings

Apart from the Git settings, you also need to configure Docker.

* **Check WSL2 is set up and running**: WSL2 provides an entire Linux kernel that allows Docker to run containers without having to manage any Virtual Machines.
With WSL2, you can use a native Linux terminal and might want to develop inside that distribution.
We do not recommend doing that to avoid any network-related issues.
* **Disk space**: If you have two hard drives, consider moving the cache to the biggest one, as Docker often requires a lot of disk space.
It is used for cache, docker images, and any volumes that your container would need.
To stop cleaning all the caches every week or so, I recommend moving all docker caches to the bigger hard drive if you have one.
* **Memory settings and swap**: By default, Docker uses all system resources available, resulting in insufficient RAM on your system.
We should set up memory and swap parameters by hand to prevent that.

### Setting up the environment

We should have installed git, Golang, and Docker for Windows by now.

To start contributing, we will need the FerretDB source code located [here](https://github.com/FerretDB/FerretDB.git).
So we must fork the repository to make it possible to send PR.

![fork the FerretDB repository](../static/img/blog/image6.png)

You need to set up old branch removal to keep your repository clean.
You need to set up the “Automatically delete head branches” flag in the “General” section of your fork repository settings.

![delete head branches](../static/img/blog/image5.png)

In your local terminal, clone the forked repository with the following command:

```js
git clone https://github.com/{username}/FerretDB.git
```

Once the cloning is complete, navigate to the source code folder via

```js
cd FerretDB
```

While working on the project code locally, it’s crucial to synchronize it with the upstream  repository and push changes to the forked repository.
To do so, we need to set up remotes.
To check if the local repository is linked with the upstream one, enter this command:

```js
git remote -v
```

If you don’t see your remote repository called origin, you can easily add it with the following command:

```js
 git remote add origin https://github.com/{username}/FerretDB.git
```

After that, we need to add the upstream project repository by running this in your terminal:

```js
git remote add upstream https://github.com/FerretDB/FerretDB.git
```

Now that we are all set with git, we can start FerretDB locally.
In FerretDB, we use the “task” tool to run every command.
To install this tool, proceed with the following steps:

```js
cd tools
go generate -x
```

That command should install the “task” tool and some other helpful utilities.
Installed tools will reside in **FerretDB/bin** directory.
Once the “task” tool is installed, we can set up the local environment.

You can list all available commands with

```js
bin\task -l
```

To run the local environment, we should simply execute

```js
bin\task env-up
```

to run the local environment.

![run the local environment](../static/img/blog/image7.png)

After that, we will have a console with FerretDB containers logs output.

![FerretDB containers](../static/img/blog/image2.png)

In a separate console window, we need to run FerretDB with

```js
bin\task run
```

![run FerretDB](../static/img/blog/image3.png)

We can open another terminal window and run tests (*bin\task test*) or “mongosh” (*bin\task mongosh*).

![terminal window](../static/img/blog/image1-1.png)

## Start contributing to FerretDB on Windows

For Windows users, starting your contribution journey to FerretDB might feel a bit challenging if it’s not correctly set up.
We’ve covered all the known Windows-specific issues you might encounter in this article when contributing to FerretDB on Windows.

An excellent place to start contributing is to select any issue labeled as the [good first issue](https://github.com/FerretDB/FerretDB/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).
Not to mention, we have an awesome community [Slack group where you can connect!](https://join.slack.com/t/ferretdb/shared_invite/zt-zqe9hj8g-ZcMG3~5Cs5u9uuOPnZB8~A)  And if you experience any other issues that aren't covered in this article or have any questions, please feel free to reach out to us on Slack or GitHub Discussions.

With your Windows environment configured and set up, read our contributing guide to start contributing to FerretDB on Windows.

(Cover photo by [Max Duzij](https://unsplash.com/es/@max_duz?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText) on [Unsplash](https://unsplash.com/s/photos/computer?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText) )
