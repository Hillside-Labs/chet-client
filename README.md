# Chet is a friendly big brother

See what is slowing you down by measuring your commands.

[Find out more](https://chet-server.fly.dev)

## What does it do?

Chet times how long your command runs that prefix with `chet`. It then stores how long the command took to run along with some other info such as your OS, git repo info, and if you're running the command in a container. Chet stores the info in SQLite DB so you can find places to speed up your workflow.

1. [Signup](https://chet-server.fly.dev) to get a client_id and client_secret
2. `make chet`
3. `./chet config -c $client_id -s $client_secret -a https://chet-server.fly.dev`
