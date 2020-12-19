import os
from fabric.api import *

env.hosts = ['kullo2.kullo.net']
env.user = 'root'

KULLOSERVER_DIR = '/opt/kulloserver'

@task(default=True)
def deploy():
	local('make')
	#TODO run tests
	with cd(KULLOSERVER_DIR):
		execute(update_preregistrations)
		execute(update_hooks)
		execute(update_message_templates)
		put('kulloserver', 'kulloserver-new', mode=0755)
		run('rm kulloserver-old', warn_only=True)
		run('mv kulloserver kulloserver-old', warn_only=True)
		run('mv kulloserver-new kulloserver')
		run('systemctl stop kulloserver', warn_only=True)
		#TODO migrations
		run('systemctl start kulloserver')

@task
def rollback():
	with cd(KULLOSERVER_DIR):
		run('mv kulloserver kulloserver-new && mv kulloserver-old kulloserver')
		run('systemctl stop kulloserver', warn_only=True)
		#TODO migrations
		run('systemctl start kulloserver')

@task
def update_preregistrations():
	with cd(KULLOSERVER_DIR):
		put('config/preregistrations.csv', 'config/preregistrations.csv')

@task
def update_hooks():
	with cd(KULLOSERVER_DIR):
		put('config/hooks', 'config')
		run('chmod +x config/hooks/*')

@task
def update_message_templates():
	with cd(KULLOSERVER_DIR):
		put('config/message_templates', 'config')

@task
def update_goose():
	gopath = os.environ['GOPATH']

	local('make goose')
	with cd(KULLOSERVER_DIR):
		put(gopath + '/bin/goose', 'goose', mode=0755)
