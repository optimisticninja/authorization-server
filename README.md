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

# Running

This will run the server on port 14000 by default over HTTPS using QUIC/served on both TCP/UDP using the self-signed cert from [cert.sh](cert.sh).

```bash
$ docker-compose up -d                           # Run PostgreSQL
$ ./scripts/psql.sh sql/V0.1.0__auth_initial.sql # Create tables
$ ./scripts/cert.sh                              # Generate self-signed certificate
$ go build ./                                    # Build project
$ ./auth -tcp                                    # Run binary and host on TCP/UDP
```

# Endpoints

If trying to CURL these endpoints, you will need a version of CURL that supports HTTP3.

- `/authorize`
- `/token`

