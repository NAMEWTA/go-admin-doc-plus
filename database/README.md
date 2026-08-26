# Database Engineering

`database/` is the root owner for database fixtures, dialect verification and operator-facing
migration documentation. Executable migrations remain module-owned source code and are invoked
only through the public `task migrate PROFILE=server-sqlite|server-postgres` command.

SQLite commands use the repository development configuration. PostgreSQL commands require the
caller to declare `GO_ADMIN_CONFIG_FILE`; scripts do not search for or print credentials. Local
database files, credentials and generated migration output are not repository assets.

The Greenfield database structure is implemented by later architecture Tickets. Nothing in this
directory promises compatibility with the legacy schema or data.
