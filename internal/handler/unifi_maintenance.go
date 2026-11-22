package handler

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

func unifiCommandEnabled(command string) error {
	if os.Getenv("FERRETDB_UNIFI_COMPAT") != "true" {
		return mongoerrors.New(mongoerrors.ErrCommandNotFound, "no such command: "+command)
	}
	return nil
}

// Bare fsync is used after the controller's collection migrations. A PostgreSQL
// checkpoint flushes dirty data through the existing durable WAL architecture.
// MongoDB's global fsync write-lock mode is not fabricated.
func (h *Handler) msgUniFiFsync(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	if err := unifiCommandEnabled("fsync"); err != nil {
		return nil, err
	}
	doc := req.Document()
	if lock := doc.Get("lock"); lock != nil && lock != false {
		return nil, mongoerrors.New(mongoerrors.ErrNotImplemented, "Global fsync write locks are not supported")
	}
	if async := doc.Get("async"); async != nil && async != false {
		return nil, mongoerrors.New(mongoerrors.ErrNotImplemented, "Asynchronous fsync is not supported")
	}
	if err := h.p.WithConnContext(ctx, func(conn *pgx.Conn) error { _, e := conn.Exec(ctx, "CHECKPOINT"); return e }); err != nil {
		return nil, err
	}
	return middleware.ResponseDoc(req, wirebson.MustDocument("ok", float64(1)))
}

func (h *Handler) msgUniFiSetFCV(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	if err := unifiCommandEnabled("setFeatureCompatibilityVersion"); err != nil {
		return nil, err
	}
	version, err := getRequiredParam[string](req.Document(), "setFeatureCompatibilityVersion")
	if err != nil {
		return nil, err
	}
	// This matched DocumentDB build exposes the fixed 7.0 compatibility level.
	// Restating it is safe; pretending to upgrade/downgrade the engine is not.
	if version != "7.0" {
		return nil, mongoerrors.New(mongoerrors.ErrBadValue, "This backend supports feature compatibility version 7.0")
	}
	return middleware.ResponseDoc(req, wirebson.MustDocument("ok", float64(1)))
}

func (h *Handler) msgUniFiSetParameter(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	if err := unifiCommandEnabled("setParameter"); err != nil {
		return nil, err
	}
	doc := req.Document()
	found := false
	var rate *int64
	for _, field := range doc.FieldNames() {
		switch field {
		case "setParameter", "$db", "lsid", "comment", "maxTimeMS", "$readPreference":
			continue
		case "tcmallocMaxTotalThreadCacheBytes":
			if _, err := cappedInteger(doc, field, false); err != nil {
				return nil, err
			}
			found = true
		case "tcmallocReleaseRate":
			value, err := cappedInteger(doc, field, false)
			if err != nil {
				return nil, err
			}
			rate = &value
			found = true
		default:
			return nil, mongoerrors.New(mongoerrors.ErrBadValue, "Unsupported runtime parameter")
		}
	}
	if !found {
		return nil, mongoerrors.New(mongoerrors.ErrBadValue, "No runtime parameter supplied")
	}
	result := wirebson.MustDocument("ok", float64(1))
	// The obsolete thread-cache cap is accepted as a compatibility hint, as on
	// the MongoDB reference. PostgreSQL has no per-thread tcmalloc cache.
	if rate != nil {
		previous := h.compatReleaseRate.Swap(*rate)
		if err := result.Add("was", previous); err != nil {
			return nil, err
		}
		// Interpret the legacy release setting as a memory-reclamation hint,
		// not as a claim that PostgreSQL implements tcmalloc's release policy.
		h.p.ReleaseIdleMemory()
		debug.FreeOSMemory()
	}
	return middleware.ResponseDoc(req, result)
}
