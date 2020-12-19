#!/bin/bash

if [ -f ~/.pgpass ]; then
	export PGPASS=$(grep "^localhost:15432:\*:kullo:" < ~/.pgpass | cut -d: -f5)
fi

"$GOPATH"/bin/goose -path config "$@"

