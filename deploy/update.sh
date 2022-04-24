#!/bin/bash

if [ $# != 2 ]; then
	echo 'need two arguments'
	return
fi

# arguments
# $1 is domain (example.com)
domain=${1}
# $2 is version from github (v2.1)
release_version=${2}


# download release
echo 'Download release...'
wget -O /usr/local/bin/frontend https://github.com/cnblvr/puzzles/releases/download/${release_version}/frontend
chmod +x /usr/local/bin/frontend
service_frontend=${domain}'-frontend'
systemctl restart ${service_frontend}
wget -O /usr/local/bin/generator https://github.com/cnblvr/puzzles/releases/download/${release_version}/generator
chmod +x /usr/local/bin/generator
service_generator=${domain}'-generator'
systemctl restart ${service_generator}
