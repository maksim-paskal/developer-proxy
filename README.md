# Proxy local requests to cluster

## Linux
```bash
docker pull paskalmaksim/developer-proxy:latest

docker run --rm -it --net=host paskalmaksim/developer-proxy:latest \
-endpoint="https://github.com" \
-rule="/graphql@http://127.0.0.1:4001" \
-rule="equal:/css/main.css@endpoint" \
-rule="regexp:^/(css|scripts)@http://127.0.0.1:4003" \
-rule="prefix:/payment@http://127.0.0.1:4004" \
```

proxy will start <http://127.0.0.1:10000>

## Proxy routing results

```bash
# if no rules match
http://127.0.0.1:10000 => https://github.com

# -rule="/graphql@http://127.0.0.1:4001"
http://127.0.0.1:10000/graphql => http://127.0.0.1:4001

# -rule="equal:/css/main.css@endpoint"
http://127.0.0.1:10000/css/main.css => https://github.com/css/main.css

# -rule="regexp:^/(css|scripts)@http://127.0.0.1:4003"
http://127.0.0.1:10000/css/test.css => http://127.0.0.1:4003/css/test.css

# -rule="regexp:^/(css|scripts)@http://127.0.0.1:4004"
http://127.0.0.1:10000/payment/form => http://127.0.0.1:4003/payment/form
```