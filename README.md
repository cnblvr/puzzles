V.2

Demo: [stuffy.space](https://stuffy.space/)

![/screenshot.png](/screenshot.png)

Unlike `V.1`, the generator creates puzzles with an emphasis on logical strategies. This version implements strategies:
* easy: Naked Single;
* normal: Naked Pair/Triple, Hidden Single/Pair/Triple;
* hard: Pointing Pair/Triple, Box-Line Reduction Pair/Triple;
* harder: X-Wing.

Controls.
* The `[c]` key on your keyboard and the `(C)` button on the screen toggles the answer/candidate entry mode. Candidates can also be entered by holding down the `[Shift]` key.
* The `[Backspace]`/`[Space]`/`[0]` keys or the `(⨯)` button clear the answer. In the "candidate input" mode, candidates are cleared.
* Keys `[1]`-`[9]` or buttons `(1)`-`(9)` put a number depending on the mode. 
* The `(h)` button suggests a possible strategy. **BUG**: the assistant focuses on your candidates, so keep them without mistakes.
* The `use highlights`, `show candidates` and `show wrongs` checkboxes make it easier to find a solution. The first turns on the highlight for the selected digit. The second shows or hides the candidates. The third shows or hides your current mistakes.

1. Create redis config and development environments
```shell
rm -f redis.conf && touch redis.conf
rm -f dev.env && touch dev.env
echo 'SEC_COOKIE_HASH_KEY='$(head -c 32 /dev/random | base64) >> dev.env
echo 'SEC_COOKIE_BLOCK_KEY='$(head -c 32 /dev/random | base64) >> dev.env
echo 'REDIS_ADDRESS=redis:6379' >> dev.env
echo 'REDIS_USER_DB=0' >> dev.env
echo 'REDIS_PUZZLE_DB=0' >> dev.env
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

3. The `generator` service will start generating puzzles of varying difficulty (10-15 minutes for the `harder` difficulty level). Open [localhost:8080](http://localhost:8080).