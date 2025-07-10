#!/bin/sh
# add ferretdb user/group (if it doesn't already exist)
err_already_exists=9
/usr/sbin/useradd --system --home-dir /nonexistent --no-create-home --shell /usr/sbin/nologin --user-group ferretdb
err=$?
if [ "$err" -ne 0 ] && [ "$err" -ne "$err_already_exists" ]; then
	exit $err
fi