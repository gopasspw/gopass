# Using gopass as summon provider

## What is summon?

[Summon](https://cyberark.github.io/summon) is a command-line tool to inject secrets as environment variables. 
It is used to execute a process and inject secrets from a separate store. Using gopass can be useful in (local) development

## Summon Provider

The gopass repository contains the [cmd/gopass-summon-provider](../cmd/gopass-summon-provider) tool, that can be used as [summon provider](https://cyberark.github.io/summon/#providers).

To make use of gopass-summon-provider to retrieve the `test/db-password` secret, you can call summon with full provider path

    summon -p /usr/local/bin/gopass-summon-provider --yaml 'DBPASS: !var test/db-password' bash -c 'echo $DBPASS'

or link gopass-summon-provider to `/usr/local/lib/summon/gopass` and just use `gopass`

    summon -p gopass --yaml 'DBPASS: !var test/db-password' bash -c 'echo $DBPASS'

or export `SUMMON_PROVIDER=gopass` as default provider

    summon --yaml 'DBPASS: !var test/db-password' bash -c 'echo $DBPASS'

With the appropriate `secrets.yml`, it's as easy as running

    summon ./my-command-to-run
