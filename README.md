# gron

*SystemD and cron inspired job scheduler*

## Usage

1. Create `gron.d` directory
2. Create job config in `gron.d/job1.conf` ([TOML](https://en.wikipedia.org/wiki/TOML) format):
    ```toml
    Type                   = "cmd"        # command execution
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
    OnErrorCmd             = "echo 'Error occurred: $ErrorText'" # execute cmd on job error
    ```
3. Launch `gron` binary
4. HTTP interface available on http://127.0.0.1:9876
