# That's a shim for https://github.com/FerretDB/github-actions/blob/main/linters/action.yml
# TODO Remove this file when https://github.com/FerretDB/dance/issues/75 is done
# and github-actions are updated to use bin/task.

init:
	cd tools; go generate -x
	bin/task init

%:
	# Use `bin/task $@` instead.
	bin/task $@
