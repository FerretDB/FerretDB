package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"slices"
	"time"

	p "github.com/SAP/go-hdb/driver/internal/protocol"
)

// check if statements implements all required interfaces.
var (
	_ driver.Stmt              = (*stmt)(nil)
	_ driver.StmtExecContext   = (*stmt)(nil)
	_ driver.StmtQueryContext  = (*stmt)(nil)
	_ driver.NamedValueChecker = (*stmt)(nil)
)

type stmt struct {
	conn  *conn
	query string
	pr    *prepareResult
	// rows: stored procedures with table output parameters
	rows *sql.Rows
}

type totalRowsAffected int64

func (t *totalRowsAffected) add(r driver.Result) {
	if r == nil {
		return
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return
	}
	*t += totalRowsAffected(rows)
}

func newStmt(conn *conn, query string, pr *prepareResult) *stmt {
	conn.metrics.msgCh <- gaugeMsg{idx: gaugeStmt, v: 1} // increment number of statements.
	return &stmt{conn: conn, query: query, pr: pr}
}

/*
NumInput differs dependent on statement (check is done in QueryContext and ExecContext):
- #args == #param (only in params):    query, exec, exec bulk (non control query)
- #args == #param (in and out params): exec call
- #args == 0:                          exec bulk (control query)
- #args == #input param:               query call.
*/
func (s *stmt) NumInput() int { return -1 }

func (s *stmt) Close() error {
	c := s.conn

	c.metrics.msgCh <- gaugeMsg{idx: gaugeStmt, v: -1} // decrement number of statements.

	if s.rows != nil {
		s.rows.Close()
	}
	if c.isBad.Load() {
		return driver.ErrBadConn
	}
	return c.dropStatementID(context.Background(), s.pr.stmtID)
}

// CheckNamedValue implements NamedValueChecker interface.
func (s *stmt) CheckNamedValue(nv *driver.NamedValue) error {
	// conversion is happening as part of the exec, query call
	return nil
}

func (s *stmt) QueryContext(ctx context.Context, nvargs []driver.NamedValue) (driver.Rows, error) {
	if s.pr.isProcedureCall() {
		return nil, fmt.Errorf("invalid procedure call %s - please use Exec instead", s.query)
	}

	c := s.conn
	c.sqlTracer.begin()

	done := make(chan struct{})
	var rows driver.Rows
	c.wg.Add(1)
	var sqlErr error
	go func() {
		defer c.wg.Done()
		rows, sqlErr = c.query(ctx, s.pr, nvargs, !s.conn.inTx)
		close(done)
	}()

	select {
	case <-ctx.Done():
		c.isBad.Store(true)
		ctxErr := ctx.Err()
		c.sqlTracer.log(ctx, traceQuery, s.query, ctxErr, nvargs)
		return nil, ctxErr
	case <-done:
		c.sqlTracer.log(ctx, traceQuery, s.query, sqlErr, nvargs)
		return rows, sqlErr
	}
}

func (s *stmt) ExecContext(ctx context.Context, nvargs []driver.NamedValue) (driver.Result, error) {
	c := s.conn
	c.sqlTracer.begin()

	if hookFn, ok := ctx.Value(connHookCtxKey).(connHookFn); ok {
		hookFn(c, choStmtExec)
	}

	done := make(chan struct{})
	var result driver.Result
	c.wg.Add(1)
	var sqlErr error
	go func() {
		defer c.wg.Done()
		if s.pr.isProcedureCall() {
			result, s.rows, sqlErr = s.execCall(ctx, s.pr, nvargs)
		} else {
			result, sqlErr = s.execDefault(ctx, nvargs)
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		c.isBad.Store(true)
		ctxErr := ctx.Err()
		c.sqlTracer.log(ctx, traceExec, s.query, ctxErr, nvargs)
		return nil, ctxErr
	case <-done:
		c.sqlTracer.log(ctx, traceExec, s.query, sqlErr, nvargs)
		return result, sqlErr
	}
}

func (s *stmt) execCall(ctx context.Context, pr *prepareResult, nvargs []driver.NamedValue) (driver.Result, *sql.Rows, error) {
	c := s.conn
	defer c.addSQLTimeValue(time.Now(), sqlTimeCall)

	callArgs, err := convertCallArgs(pr.parameterFields, nvargs, c.attrs._cesu8Encoder(), c.attrs._lobChunkSize)
	if err != nil {
		return nil, nil, err
	}
	inputParameters, err := p.NewInputParameters(callArgs.inFields, callArgs.inArgs)
	if err != nil {
		return nil, nil, err
	}
	if err := c.write(ctx, c.sessionID, p.MtExecute, false, (*p.StatementID)(&pr.stmtID), inputParameters); err != nil {
		return nil, nil, err
	}

	/*
		call without lob input parameters:
		--> callResult output parameter values are set after read call
		call with lob output parameters:
		--> callResult output parameter values are set after last lob input write
	*/

	cr, ids, numRow, err := c.execCall(ctx, callArgs.outFields)
	if err != nil {
		return nil, nil, err
	}

	if len(ids) != 0 {
		/*
			writeLobParameters:
			- chunkReaders
			- cr (callResult output parameters are set after all lob input parameters are written)
		*/
		if err := c.writeLobs(ctx, cr, ids, callArgs.inFields, callArgs.inArgs); err != nil {
			return nil, nil, err
		}
	}

	numOutputField := len(cr.outputFields)
	// no output fields -> done
	if numOutputField == 0 {
		return driver.RowsAffected(numRow), nil, nil
	}

	scanArgs := make([]any, numOutputField)
	for i := 0; i < numOutputField; i++ {
		scanArgs[i] = callArgs.outArgs[i].Value.(sql.Out).Dest
	}

	// no table output parameters -> QueryRow
	if len(callArgs.outFields) == len(callArgs.outArgs) {
		if err := stdConnTracker.callDB().QueryRow("", cr).Scan(scanArgs...); err != nil {
			return nil, nil, err
		}
		return driver.RowsAffected(numRow), nil, nil
	}

	// table output parameters -> Query (needs to kept open)
	rows, err := stdConnTracker.callDB().Query("", cr)
	if err != nil {
		return nil, rows, err
	}
	if !rows.Next() {
		return nil, rows, rows.Err()
	}
	if err := rows.Scan(scanArgs...); err != nil {
		return nil, rows, err
	}
	return driver.RowsAffected(numRow), rows, nil
}

func (s *stmt) execDefault(ctx context.Context, nvargs []driver.NamedValue) (driver.Result, error) {
	c := s.conn

	numNVArg, numField := len(nvargs), s.pr.numField()

	if numNVArg == 0 {
		if numField != 0 {
			return nil, fmt.Errorf("invalid number of arguments %d - expected %d", numNVArg, numField)
		}
		return c.exec(ctx, s.pr, nvargs, !c.inTx, 0)
	}
	if numNVArg == 1 {
		if _, ok := nvargs[0].Value.(func(args []any) error); ok {
			return s.execFct(ctx, nvargs)
		}
	}
	if numNVArg == numField {
		return s.exec(ctx, s.pr, nvargs, !c.inTx, 0)
	}
	if numNVArg%numField != 0 {
		return nil, fmt.Errorf("invalid number of arguments %d - multiple of %d expected", numNVArg, numField)
	}
	return s.execMany(ctx, nvargs)
}

// ErrEndOfRows is the error to be returned using a function based bulk exec to indicate
// the end of rows.
var ErrEndOfRows = errors.New("end of rows")

/*
Non 'atomic' (transactional) operation due to the split in packages (bulkSize),
execMany data might only be written partially to the database in case of hdb stmt errors.
*/
func (s *stmt) execFct(ctx context.Context, nvargs []driver.NamedValue) (driver.Result, error) {
	c := s.conn

	totalRowsAffected := totalRowsAffected(0)
	args := make([]driver.NamedValue, 0, s.pr.numField())
	scanArgs := make([]any, s.pr.numField())

	fct, ok := nvargs[0].Value.(func(args []any) error)
	if !ok {
		panic("should never happen")
	}

	done := false
	batch := 0
	for !done {
		args = args[:0]
		for i := 0; i < c.attrs._bulkSize; i++ {
			err := fct(scanArgs)
			if errors.Is(err, ErrEndOfRows) {
				done = true
				break
			}
			if err != nil {
				return driver.RowsAffected(totalRowsAffected), err
			}

			args = slices.Grow(args, len(scanArgs))
			for j, scanArg := range scanArgs {
				nv := driver.NamedValue{Ordinal: j + 1}
				if t, ok := scanArg.(sql.NamedArg); ok {
					nv.Name = t.Name
					nv.Value = t.Value
				} else {
					nv.Name = ""
					nv.Value = scanArg
				}
				args = append(args, nv)
			}
		}

		r, err := s.exec(ctx, s.pr, args, !c.inTx, batch*c.attrs._bulkSize)
		totalRowsAffected.add(r)
		if err != nil {
			return driver.RowsAffected(totalRowsAffected), err
		}
		batch++
	}
	return driver.RowsAffected(totalRowsAffected), nil
}

/*
Non 'atomic' (transactional) operation due to the split in packages (bulkSize),
execMany data might only be written partially to the database in case of hdb stmt errors.
*/
func (s *stmt) execMany(ctx context.Context, nvargs []driver.NamedValue) (driver.Result, error) {
	c := s.conn
	bulkSize := c.attrs._bulkSize

	totalRowsAffected := totalRowsAffected(0)
	numField := s.pr.numField()
	numNVArg := len(nvargs)
	numRec := numNVArg / numField
	numBatch := numRec / c.attrs._bulkSize
	if numRec%c.attrs._bulkSize != 0 {
		numBatch++
	}

	for i := 0; i < numBatch; i++ {
		from := i * numField * bulkSize
		to := (i + 1) * numField * bulkSize
		if to > numNVArg {
			to = numNVArg
		}
		r, err := s.exec(ctx, s.pr, nvargs[from:to], !c.inTx, i*bulkSize)
		totalRowsAffected.add(r)
		if err != nil {
			return driver.RowsAffected(totalRowsAffected), err
		}
	}
	return driver.RowsAffected(totalRowsAffected), nil
}

/*
exec executes a sql statement.

Bulk insert containing LOBs:
  - Precondition:
    .Sending more than one row with partial LOB data.
  - Observations:
    .In hdb version 1 and 2 'piecewise' LOB writing does work.
    .Same does not work in case of geo fields which are LOBs en,- decoded as well.
    .In hana version 4 'piecewise' LOB writing seems not to work anymore at all.
  - Server implementation (not documented):
    .'piecewise' LOB writing is only supported for the last row of a 'bulk insert'.
  - Current implementation:
    One server call in case of
    .'non bulk' execs or
    .'bulk' execs without LOBs
    else potential several server calls (split into packages).
  - Package invariant:
    .for all packages except the last one, the last row contains 'incomplete' LOB data ('piecewise' writing)
*/
func (s *stmt) exec(ctx context.Context, pr *prepareResult, nvargs []driver.NamedValue, commit bool, ofs int) (driver.Result, error) {
	c := s.conn
	defer c.addSQLTimeValue(time.Now(), sqlTimeExec)

	addLobDataRecs, err := convertExecArgs(pr.parameterFields, nvargs, c.attrs._cesu8Encoder(), c.attrs._lobChunkSize)
	if err != nil {
		return driver.ResultNoRows, err
	}

	// piecewise LOB handling
	numColumn := len(pr.parameterFields)
	totalRowsAffected := totalRowsAffected(0)
	from := 0
	for i := 0; i < len(addLobDataRecs); i++ {
		to := (addLobDataRecs[i] + 1) * numColumn

		r, err := c.exec(ctx, pr, nvargs[from:to], commit, ofs)
		totalRowsAffected.add(r)
		if err != nil {
			return driver.RowsAffected(totalRowsAffected), err
		}
		from = to
	}
	return driver.RowsAffected(totalRowsAffected), nil
}
