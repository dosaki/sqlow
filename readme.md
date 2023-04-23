<div align="center"><img src="./icon.svg" /></div>
<h1 align="center">Sqlow</h1>

_**pronounced Squallow** - a mix of Swallow (the migratory bird) and SQL_

A CLI database schema migrator that doesn't care about versions.

Supported databases:
* MariaDB / Mysql
* PostgreSQL

# Usage

Create a file like below:
```yaml
# migration.yml

migrations:
  - description: Ensure that description exists for task_types
    check: SELECT description FROM task_types;
    onFail:
      - ALTER TABLE task_types ADD COLUMN description VARCHAR(1023);
      - UPDATE task_types SET description = CONCAT('Task type for ', name);
      - ALTER TABLE task_types ALTER COLUMN description SET NOT NULL;
```

Then run it!
```shell
sqlow -pSup4rS@f3P@ssw0rd run ./migration.yml
```


## Migration files
A valid migration file must start with `migrations:` with an array of migrations defined under it, like so:
```yaml
migrations:
  - description: What this migration is meant to do
    check: your pass/fail SQL here
    onFail: sql to run if check fails (throws any errors)
    onSuccess: sql to run if check passes
    onResults: sql to run if check has 1 or more results
    onNoResults: sql to run if check has 0 results
    always: this sql always runs
```

* `description` is always a mandatory field.
* For any of the `onXXX` fields to be valid, a `check` must be present
* `always` will run even without a `check` field
* The following fields also support a list of SQL strings:
  * `check`
  * `onFail`
  * `onSuccess`
  * `onResults`
  * `onNoResults`
  * `always`

These fields can all be used together... if that makes any sense for your scenario.
If that's something you're doing, the order is as listed above.

### Pointing to SQL files
You can also reference SQL files by appending `File` to any of the fields. E.g.:
```yaml
migrations:
  - description: What this migration is meant to do
    checkFile: ./path/to/your/check.file.sql
    onFail: some additional SQL to run on fail
    onFailFile: ./path/to/your/onFail.file.sql
```

Note:
* `onFail` will always run before `onFailFile`!
* `xxxFile` fields can only take a single string, not a list!

### What if I still want versions?
Versions might be useful if you want to run something once and never again and you might not have an easy way to check whether it's been done.

You can do it with the methods below. If this is a common scenario for you, vote on issue [#2](https://github.com/dosaki/sqlow/issues/2).

#### Actual versions
First ensure there's a table to store the info:
```yaml
migrations:
  - description: Ensure there's a migrations table
    check: SELECT * FROM pg_tables WHERE schemaname = 'public' AND tablename ='migrations';
    onNoResults:
      - CREATE TABLE migrations (id int primary key, version int NOT NULL);
      - INSERT INTO migrations (id, version) VALUES (1, 0);
```

... and then write your migrations:
```yaml
migrations:
  - description: Do something once
    check: select * from migrations where id = 1 and version < 1;
    onResults:
      - Some SQL here that you want to run once
      - UPDATE migrations SET version = 1 WHERE id = 1;
  - description: Do another thing once
    check: select * from migrations where id = 1 and version < 2;
    onResults:
      - Another SQL here that you want to run once
      - UPDATE migrations SET version = 2 WHERE id = 1;
```

#### Using something a little more descriptive
First ensure there's a table to store the info:
```yaml
migrations:
  - description: Ensure there's a migrations table
    check: SELECT * FROM pg_tables WHERE schemaname = 'public' AND tablename ='migrations';
    onNoResults:
      - CREATE TABLE migrations (description varchar(255) NOT NULL);
```

... and then write your migrations:
```yaml
migrations:
  - description: Do something once
    check: select * from migrations where description = 'Do something once';
    onNoResults:
      - Some SQL here that you want to run once
      - INSERT into migrations SET description = 'Do something once';
  - description: Do another thing once
    check: select * from migrations where description = 'Do another thing once';
    onNoResults:
      - Another SQL here that you want to run once
      - INSERT into migrations SET description = 'Do another thing once';
```

I prefer this way since it acts as a nice human-readable log of which single-run migrations have taken effect.

## Help

```
Usage:
  sqlow [options] -p<password> run <path>

Options:
  -h --help                 Show this Help.
  -r --recursive            Traverse the directory structure for .yml or .yaml files.
  -c --config=<configPath>  Path to config [default: ./config.yml].
  -e --engine=<engine>      Database engine (overrides config).
  -h --host=<host>          Database host (overrides config).
  -P --port=<port>          Database port (overrides config).
  -s --schema=<schema>      Database schema (overrides config).
  -u --username=<username>  Database username (overrides config).
  -p --password=<password>  Database password (required).
  -o --options=<opts>       Comma separated list of key:value options (merges with config).

Examples:
  sqlow -r -pSup4rS@f3P@ssw0rd run ./migrations
  sqlow -pSup4rS@f3P@ssw0rd run ./migrations/a_migration.yml
  sqlow -c ./my-config.yml -pSup4rS@f3P@ssw0rd run ./migrations/a_migration.yml
  sqlow -r -c ./my-config.yml -udbuser -pSup4rS@f3P@ssw0rd run ./migrations/
```

# Configuration

You can configure the target database with a config file like below:
```yaml
# config.yml

engine: postgres   # or maria or mysql
host: your.host    # host address
port: 12345        # port for the host
schema: sqlow      # database schema
username: sqlow    # username to connect to the database. You'll likely need something with full privileges (passwords must be provided via CLI)
options:           # any options for the connection string
  option1: value1 
  option2: value2 
```
Notes:
* Any CLI item will override the configurations in the file.
* Password must be passed in via the CLI (`-pSup4rS@f3P@ssw0rd`)
