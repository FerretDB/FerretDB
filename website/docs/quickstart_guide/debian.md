---
sidebar_position: 2
---

# Debian package

To install the .deb packages for FerretDB on your Debian, Ubuntu, Linux, and other Unix-like systems, you can use `apt` or `dpkg`. 

Download the latest FerretDB .deb package from [our release pages](https://github.com/FerretDB/FerretDB/releases).

To install FerretDB .deb package using `dpkg`, copy the run the following command in your terminal:

```sh
$ sudo apt update
# Install using dpkg
$ sudo dpkg -i <filename>.deb
```

It’s important to note that `dpkg` doesn’t address dependencies. If you encounter any errors related to dependencies when installing, you can resolve them using the following command:

```sh
sudo apt install -f
```

Instead of using `dpkg`, you can use apt to manage, install, and resolve all package dependencies automatically.
However, you must specify the full path to the .deb file.
This will stop `apt` from downloading and installing from Ubuntu’s repositories.

In the root repository of the debian file, you should indicate the path by adding "`./`".

```sh
$ sudo apt update
# Install via apt
$ sudo apt install ./<filename>.deb
```
Once FerretDB is installed, you can start the software.