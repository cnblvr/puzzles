1. Create redis config and development environments
```shell
rm -f redis.conf && touch redis.conf
rm -f dev.env && touch dev.env
echo 'SEC_COOKIE_HASH_KEY='$(head -c 32 /dev/random | base64) >> dev.env
echo 'SEC_COOKIE_BLOCK_KEY='$(head -c 32 /dev/random | base64) >> dev.env
echo 'REDIS_ADDRESS=redis:6379' >> dev.env
echo 'REDIS_USER_DB=0' >> dev.env
echo 'PASSWORD_PEPPER='$(head -c 32 /dev/random | base64) >> dev.env
# optional: set password for redis
export REDISPASSWORD=$(head -c 16 /dev/random | base64)
echo "requirepass $REDISPASSWORD" >> redis.conf
echo "REDIS_PASSWORD=$REDISPASSWORD" >> dev.env
```

2. Run this application
```shell
sudo docker-compose up --build
```