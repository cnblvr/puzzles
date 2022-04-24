#!/bin/bash

if [ $# != 3 ]; then
	echo 'need three arguments'
	return
fi

# arguments
# $1 is domain (example.com)
domain=${1}
# $2 is email for Let's Encrypt (example@email.com)
email=${2}
# $3 is version from github (v2)
release_version=${3}

redis_password=$(head -c 16 /dev/random | base64)


apt -y update


# install nginx
echo 'Install nginx...'
apt -y install curl gnupg2 ca-certificates lsb-release ubuntu-keyring
curl https://nginx.org/keys/nginx_signing.key | gpg --dearmor \
    | sudo tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null
gpg --dry-run --quiet --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/ubuntu `lsb_release -cs` nginx" \
    | tee /etc/apt/sources.list.d/nginx.list
apt -y update
apt -y install nginx


# install redis
echo 'Install redis...'
curl -fsSL https://packages.redis.io/gpg \
    | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" \
    | tee /etc/apt/sources.list.d/redis.list
apt -y update
apt -y install redis
printf '
requirepass '${redis_password}'
' >> /etc/redis/redis.conf
systemctl restart redis-server.service


# configure .env file
echo 'Configure .env file...'
dot_env_filename=/usr/local/etc/${domain}.env
printf 'SEC_COOKIE_HASH_KEY='$(head -c 32 /dev/random | base64)'
SEC_COOKIE_BLOCK_KEY='$(head -c 32 /dev/random | base64)'
REDIS_ADDRESS=localhost:6379
REDIS_USER_DB=0
REDIS_PUZZLE_DB=0
PASSWORD_PEPPER='$(head -c 32 /dev/random | base64)'
REDIS_PASSWORD='${redis_password}'
' > ${dot_env_filename}
echo 'Env filename: '${dot_env_filename}


# download release
echo 'Download release...'
wget -O /usr/local/bin/frontend https://github.com/cnblvr/puzzles/releases/download/${release_version}/frontend
chmod +x /usr/local/bin/frontend
service_frontend=${domain}'-frontend'
wget -O /usr/local/bin/generator https://github.com/cnblvr/puzzles/releases/download/${release_version}/generator
chmod +x /usr/local/bin/generator
service_generator=${domain}'-generator'
printf '[Unit]
Description=Puzzles '${release_version}' frontend service
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/frontend
Restart=always
EnvironmentFile='${dot_env_filename}'

[Install]
WantedBy=multi-user.target
' > /etc/systemd/system/${service_frontend}.service
systemctl enable ${service_frontend}
systemctl start ${service_frontend}
printf '[Unit]
Description=Puzzles '${release_version}' generator service
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/generator
Restart=always
EnvironmentFile='${dot_env_filename}'

[Install]
WantedBy=multi-user.target
' > /etc/systemd/system/${service_generator}.service
systemctl enable ${service_generator}
systemctl start ${service_generator}
echo 'View logs:
  journalctl -u '${service_frontend}' -b
  journalctl -u '${service_generator}' -b'


# configure nginx
echo 'Configure nginx...'
mv /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf.disabled
printf 'server {
    server_name '${domain}';
    location / {
        proxy_pass http://localhost:8080/;
    }
    location /game_ws {
        proxy_pass http://localhost:8080/game_ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Origin http://localhost:8080;
    }
}
' > /etc/nginx/conf.d/${domain}.conf
nginx -s reload


# install certbot for let's encrypt
echo 'Install certbot...'
apt -y install snapd
snap install core
snap refresh core
snap install --classic certbot
ln -s /snap/bin/certbot /usr/bin/certbot
certbot --nginx -m ${email} --no-eff-email --agree-tos -d ${domain}
