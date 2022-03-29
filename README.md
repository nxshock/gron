# gron

*SystemD and cron inspired job scheduler*

## Usage

1. Create `gron.d` directory
2. Create job config in `jobs.d/job1.conf` ([TOML](https://en.wikipedia.org/wiki/TOML) format):
    ```toml
    Cron        = "* * * * *"                # cron instructions
    Command     = "echo Hello"               # command to execute
    Description = "print Hello every minute" # job description
    ```
3. Launch `gron` binary
4. HTTP interface available on http://127.0.0.1:9876
