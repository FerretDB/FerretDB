#!/bin/bash

# This script should only be used in Github Actions and is not made to be used for installing locally on computer

if [ "$1" != "install" ];
then
    echo "Please do not use this script for installing HANA Go driver locally on computer. This is for Github Actions"
    exit 1
fi

# Get GOROOT
prefix="GOROOT=\""
suffix="\""
gorootStr="$(go env | grep GOROOT)"
goroot=${gorootStr#"$prefix"}
goroot=${goroot%"$suffix"}

# Get home directory and navigate to it
home_dir=$HOME
cd "${home_dir}"

# Create folder for downloading and installing the HANA Go driver
mkdir hanaDriver

curl https://tools.hana.ondemand.com/additional/hanaclient-latest-linux-x64.tar.gz -H 'Cookie: eula_3_1_agreed=tools.hana.ondemand.com/developer-license-3_1.txt'  --output hanaDriver/hanaclient.tar.gz

tar -xzvf hanaDriver/hanaclient.tar.gz -C hanaDriver

# Install HANA client
hanaDriver/client/./hdbinst --batch --ignore=check_diskspace

# Move driver to GOROOT
sudo mv sap/hdbclient/golang/src/SAP "${goroot}"/src/

cd sap/hdbclient/golang/src

# Install Go driver
go install SAP/go-hdb/driver

# Remove folder for download and installation
cd "${home_dir}"

rm -rf hanaDriver