#!/usr/bin/env python3

import csv
import psycopg2
import os
import subprocess
import sys
import yesnoquestion

# Register users:
# * test1#kullo.test
# * test2#kullo.test
users = [
    {
    	'address': 'test1#kullo.test',
    	'masterKeyFile': 'data/test1#kullo.test.kpem'
    },
    {
    	'address': 'test2#kullo.test',
    	'masterKeyFile': 'data/test2#kullo.test.kpem'
    }
]

# Config
pgHost = "127.0.0.1"
pgUser = "kullo"
pgPort = "5432"
pgPass = False

if not yesnoquestion.ask("This will remove all data from database 'kullo'. Continue?", default="no"):
    sys.exit(0)

# Begin environment checks
if subprocess.call("which registerer > /dev/null", shell=True) != 0:
    print("Command 'registerer' is not in PATH.")
    print("Nothing written yet.")
    sys.exit(1)

with open(os.path.expanduser('~/.pgpass'), newline='') as csvfile:
    pgpasswords = csv.reader(csvfile, delimiter=':')
    for row in pgpasswords:
        if row[0] == pgHost and row[1] == pgPort and row[3] == pgUser:
            pgPass = row[4]

if pgPass == False:
    print("Postgres password for user '%s' not found in ~/.pgpass" % pgUser)
    print("Nothing written yet.")
    sys.exit(1)

# Connect to an existing database
try:
    conn = psycopg2.connect(host=pgHost, port=pgPort, user=pgUser, password=pgPass, dbname="kullo")
except psycopg2.DatabaseError as e:
    print(e)
    print("Nothing written yet.")
    sys.exit(1)
# End environment checks


# Clear database
try:
    with conn.cursor() as cur:
        sql = "TRUNCATE addresses, keys_asymm, keys_symm, messages, notifications, users;"
        cur.execute(sql)
except:
    print("Could not run SQL statement")
    sys.exit(1)

conn.commit()
cur.close()
conn.close()


# Register users
if subprocess.call("which registerer > /dev/null", shell=True) != 0:
    print("Command 'registerer' is not in PATH")
    sys.exit(1)

for user in users:
    address = user['address']
    masterKeyFile = user['masterKeyFile']
    print("Registering %s ..." % address)
    subprocess.check_call('registerer "%s" "%s"' % (address, masterKeyFile),
                          shell=True)

