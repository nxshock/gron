# gron

*SystemD and cron inspired job scheduler*

## Usage

1. Create `gron.d` directory
2. Create job config in `gron.d/job1.conf` ([TOML](https://en.wikipedia.org/wiki/TOML) format):
    ```toml
    Type                   = "cmd"        # command execution
    Category               = "Test jobs"  # jobs category name
    Cron                   = "* * * * *"  # cron instructions

    Command                = "echo Hello" # command to execute
    ```

    SQL job:
    ```toml
    Type                   = "sql"                            # sql execution
    Cron                   = "* * * * *"                      # cron instructions
    Description            = "execute procedure every minute" # job description

    Driver                 = "pgx"                            # "pgx" for Postgresql, "oracle" for Oracle, "sqlserver" for Microsoft SQL Server
    ConnectionString       = "postgres://login:password@host:port/database?sslmode=disable" # each driver has different syntax
    SqlText                = "CALL procedure"                 # command to execute
    ```

    Add other options if needed:
    ```toml
    Description            = "print Hello every minute" # job description
    NumberOfRestartAttemts = 3                          # number of restart attemts
    RestartSec             = 5                          # the time to sleep before restarting a job (seconds)
    RestartRule            = "on-error"                 # Configures whether the job shall be restarted when the job process exits

    OnSuccessCmd           = "echo 'Job finished.'"              # execute cmd on job success
    OnErrorCmd             = "echo 'Error occurred: {{.Error}}'" # execute cmd on job error


	OnSuccessHttpGetUrl    = ""
	OnErrorHttpGetUrl      = "http://127.0.0.1/alerts?title={{.JobName}}%20failed&message={{.Error}}&tags=warning"

    OnSuccessHttpPostUrl   = "http://127.0.0.1/alerts"
    OnSuccessMessageFmt    = "Job {{.JobName}} finished."

    OnErrorHttpPostUrl     = "http://127.0.0.1/alerts"
    OnErrorMessageFmt      = "Job {{.JobName}} failed:\n\n{{.Error}}"
    ```
3. Launch `gron` binary
4. HTTP interface available on http://127.0.0.1:9876
