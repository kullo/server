# The Kullo Server

Kullo server written in Go.

## Requirements

* Go version 1.2
* PostgreSQL server 9.3+

## Setup GOPATH

Create a new folder where your go sources and packages will be stored:

    mkdir ~/go
    cd ~/go
    mkdir bin pkg src

Define GOPATH:

    # add this to your ~/.bashrc or ~/.zshrc
    export GOPATH="${HOME}/go"

Open a new shell or source the edited file to apply changes.

## Setup: Get sources

Initial checkout of private repository:

    cd $GOPATH/src
    mkdir -p bitbucket.org/kullo
    cd bitbucket.org/kullo
    git clone git@bitbucket.org:kullo/server.git

Get, update and build dependencies:

    cd $GOPATH/src/bitbucket.org/kullo/server
    make update
    make

## Install PostgreSQL

[Ubuntuusers Wiki: PostgreSQL](http://wiki.ubuntuusers.de/PostgreSQL)

### Install server package

Ubuntu:

    sudo apt-get install postgresql

Fedora:

    sudo yum install postgresql-server

### Create database cluster

Ubuntu:

    Cluster 'main' is created automatically.

Fedora:

    sudo postgresql-setup initdb

### Make postgres locally accessible

Edit `pg_hba.conf` (Ubuntu: `/etc/postgresql/9.3/main/`, Fedora: `/var/lib/pgsql/data/`) so that the following line is present and no other line matches the first 4 columns:

    host    all             all             127.0.0.1/32            md5

Afterwards, start postgres:

    sudo service postgresql start

### Database passwords

The current database passwords can be read from the ansible configuration:

    cd workspace/ansible
    ansible-vault view --vault-password-file=vault-password.txt roles/kullo-server/vars/main.yml

The password for the `kullo` user should be stored in `~/.pgpass`.

## Create database

Create database application user and application database

    sudo -u postgres createuser -P kullo
    sudo -u postgres createdb -O kullo kullo

Set database settings

    cd $GOPATH/src/bitbucket.org/kullo/server
    vi config/dbconf.yml


### Install goose

    go get -u bitbucket.org/liamstask/goose/cmd/goose

### Check migration status:

    cd $GOPATH/src/bitbucket.org/kullo/server
    ./goose.sh -env <environment> status

Migrate migrations if necessary.

    ./goose.sh -env <environment> up


## Start server locally

The Kullo server must be started from the right working directory in order to
find config files like `config/dbconf.yml`.

    cd $GOPATH/src/bitbucket.org/kullo/server
    make && ./kulloserver -env <environment>


## Running integration tests

Once:

    sudo -u postgres createuser -P kullotest
    sudo -u postgres createdb -O kullotest kullotest
    ./goose.sh -env integrationtest up

    virtualenv /path/to/new/venv
    source /path/to/new/venv/bin/activate
    pip install -r tests/requirements.txt

Every time, in one shell:

    make && ./kulloserver -env integrationtest

Every time, in another shell:

    source /path/to/new/venv/bin/activate  # if not yet activated in this shell
    make integrationtest

## Deploy to server

As a prerequisite, this needs the tool [fabric](http://www.fabfile.org/), which is available from the package repositories of at least Ubuntu and Fedora under the name `fabric`.

### Push a new version to the server

    fab

### Revert to the previous version

    fab rollback
