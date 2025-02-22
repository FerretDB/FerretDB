#include <dlfcn.h>
#include "postgres.h"
#include "pgstat.h"
#include "postmaster/bgworker.h"
#include "storage/ipc.h"
#include "storage/latch.h"
#include "storage/proc.h"
#include "miscadmin.h"
#include "utils/elog.h"
#include "fmgr.h"

PG_MODULE_MAGIC;

static char *ferretdb_lib = "ferretdb.so";

PGDLLEXPORT void background_main(Datum main_arg);

void background_main(Datum main_arg)
{
    bgworker_main_type entrypt;
    char *golib_path;
    void *handle;

    BackgroundWorkerUnblockSignals();

    golib_path = (char *)palloc(strlen(pkglib_path) + 1 +
                                strlen(ferretdb_lib) + 1);

    join_path_components(golib_path, pkglib_path, ferretdb_lib);
    elog(DEBUG1, "ferretdb_loader: loading go shared lib \"%s\"", golib_path);

    handle = dlopen(golib_path, RTLD_NOW | RTLD_GLOBAL);
    pfree(golib_path);

    entrypt = (bgworker_main_type)dlsym(handle, "BackgroundWorkerMain");

    entrypt(main_arg);

    dlclose(handle);

    proc_exit(2);
}

void _PG_init(void)
{
    BackgroundWorker worker;
    MemSet(&worker, 0, sizeof(BackgroundWorker));

    snprintf(worker.bgw_name, BGW_MAXLEN, "FerretDBLoader");
    worker.bgw_flags = BGWORKER_SHMEM_ACCESS;
    worker.bgw_start_time = BgWorkerStart_RecoveryFinished;
    worker.bgw_restart_time = BGW_NEVER_RESTART;
    snprintf(worker.bgw_library_name, BGW_MAXLEN, "ferretdb_loader");
    snprintf(worker.bgw_function_name, BGW_MAXLEN, "background_main");

    RegisterBackgroundWorker(&worker);
}
