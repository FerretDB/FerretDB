---
sidebar_position: 2
---

# Debian package

To install the .deb packages for FerretDB on your Debian, Ubuntu, Linux, and other Unix-like systems, you can use `apt` or `dpkg`.

Download the latest FerretDB .deb package from [our release pages](https://github.com/FerretDB/FerretDB/releases).

To install FerretDB .deb package using `dpkg`, run the following command in your terminal:

```sh
$ sudo apt update
# Install using dpkg
$ sudo dpkg -i <filename>.deb
```

It’s important to note that `dpkg` doesn’t address dependencies.
If you encounter any errors related to dependencies when installing, you can resolve them by installing all package dependencies with the following command:

```sh
sudo apt install -f
```

Instead of using `dpkg`, you can use `apt` to manage, install, and resolve all package dependencies automatically.
However, you must specify the full path to the .deb package.
This will stop `apt` from downloading and installing from Ubuntu’s repositories.

For example, if the file is in your current working directory, indicate the path by prepending "`./`" to the filename.

```sh
$ sudo apt update
# Install via apt
$ sudo apt install ./<filename>.deb
```

Once FerretDB is installed, you can start the software.
See [our Docker guide](docker.md) and [Configuration flags and variables](../configuration/flags.md) page for more details.
