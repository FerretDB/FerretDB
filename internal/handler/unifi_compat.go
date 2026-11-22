package handler

import "github.com/FerretDB/wire/wirebson"

// unifiDurabilityAlias supports UniFi Network 10.6.109's engine-specific
// journaling probe. It is an explicitly named PostgreSQL boolean, not a
// fabricated WiredTiger counter. The caller enables it only through the
// FERRETDB_UNIFI_COMPAT opt-in. Ordinary serverStatus remains unchanged.
func unifiDurabilityAlias(fsync, fullPageWrites, synchronousCommit string) *wirebson.Document {
	enabled := int32(0)
	if fsync == "on" && fullPageWrites == "on" {
		switch synchronousCommit {
		case "on", "local", "remote_write", "remote_apply":
			enabled = 1
		}
	}
	return wirebson.MustDocument("log", wirebson.MustDocument("ferretdb.postgresqlDurableWAL", enabled))
}
