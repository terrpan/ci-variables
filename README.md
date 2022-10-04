# ci-variables

Ci-variables is simple Go program to fetch all project variables in a group in Gitlab. It's used in CI-pipelines for all projects that uses Helm. This is a workaround for high permissions issues with Group variables.

## Usage

### Arguments
| Short|Long|Help|Required|
|---|---|---|---|
|-t|token|Gitlab token|True|
|-p|projectId|ID of of project|True|
|-s|scope|EnvironmentScope for ci variables|True|
|-o|dir|Output directory, if not set CWD will be used|False|
|-d|log-level|Set log level. Accepts "INFO", "DEBUG", "WARN", "ERROR"|True|

### Running the binary directly
```bash
./ci-variables -t gitlabtoken -p projectid -s staging -o /tmp/
```

### Using Docker

You can mount a directory (eg. /tmp/) to use as output directory.

```bash
docker run --rm --entrypoint "" --volume "/tmp/:/tmp/" -u $(id -u) registry.gitlab.com/made-people/utilities/ci-variables/main:latest -t gitlabtoken -p projectid -s staging -o /tmp/
```
### Gitlab CI

TODO
