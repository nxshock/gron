package main

import (
	"database/sql"
	"io"

	_ "github.com/denisenkom/go-mssqldb"
	mssql "github.com/denisenkom/go-mssqldb"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/nxshock/logwriter"
	_ "github.com/sijms/go-ora/v2"
)

// openMsSqlDb copies MS SQL Driver because logger can be assigned only for driver
func openMsSqlDb(dataSourceName string, logger io.Writer) (*sql.DB, error) {
	// Init logger
	lw := logwriter.New(logger)
	defer lw.Close()
	lw.TimeFormat = config.TimeFormat

	driverInstance := &mssql.Driver{} // TODO: check hidden processQueryText field
	driverInstance.SetLogger(&LogWriter{lw})

	connector, err := driverInstance.OpenConnector(dataSourceName)
	if err != nil {
		return nil, err
	}

	return sql.OpenDB(connector), nil
}
