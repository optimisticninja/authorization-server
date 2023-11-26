# auth

This project is an OAuth 2.0 authorization server based on a modified version [OpenShift's osin project](https://github.com/openshift/osin).

The two divergent forks of others' work can be found here:

- My [osin](https://github.com/optimisticninja/osin)
    - Argon2 hashing was added for client secrets
    - No changes were made to access/refresh tokens, as they only live for an hour - this may change in the future
    - Removal of vendor directory/dependency updates
- My [osin-postgres](https://github.com/optimisticninja/osin-postgres)
    - Simply pointed at my version of osin in the dependencies
    - Dependency modifications

## TODO

* Implement own storage for encrypted secrets
    - https://github.com/openshift/osin/blob/master/example/teststorage.go

## Config

All configuration is handled in a `.env` file at that sits in the directory where the binary is executed from.

|Config|Default Value|
|------|-------------|
|DB_HOST|127.0.0.1|
|DB_PORT|5432|
|DB_USER|root|
|DB_PASSWORD|root|
|DB|auth|
|CERT_PATH|./cert/cert.pem|
|KEY_PATH|./cert/priv.key|
|SERVER_HOST|:14000|

## Running

This will run the server on port 14000 by default over HTTPS using QUIC/served on both TCP/UDP using the self-signed cert from [cert.sh](cert.sh).

```bash
$ docker-compose up -d                           # Run PostgreSQL
$ ./scripts/psql.sh sql/V0.1.0__auth_initial.sql # Create tables
$ pushd clihasher && \
    go build . && \
    ./clihasher myClientSecret && 
    popd                                         # Create Argon2 hash for client secret/copy
$ ./scripts/psql.sh
...
auth=# INSERT INTO client VALUES('my-id', 'copied Argon2 hash from previous', '', 'https://localhost:14000/info');
auth=# \q
$ ./scripts/cert.sh                              # Generate self-signed certificate
$ go build ./                                    # Build project
$ ./auth -tcp                                    # Run binary and host on TCP/UDP
```

## Endpoints

If trying to CURL these endpoints, you will need a version of CURL that supports HTTP3.

- `/authorize`
- `/token`

## Testing in Postman

First, ensure the service is running on tcp with `./auth -tcp` and that you have inserted a client with an
Argon2 hashed secret into the `clients` table.

1. Create new OAuth 2.0 under Authorization tab.
2. Choose to add auth data to request headers or URL.
2. Set grant type to `Authorization Code (with PKCE)`.
3. Add a header prefix of `Bearer`.
4. The callback URL must match what you added to the database.
5. Auth URL is: `https://localhost:14000/authorize`
6. Access token URL is: `https://localhost:14000/token`
7. Add your created client ID/unhashed secret.
8. Choose your client authentication to be in body or Basic Auth header.



