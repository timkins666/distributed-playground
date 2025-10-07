# This is my Go/distributed systems learning sandbox

It's a somewhat chaotic, miniaturised, distributed banking implementation using Dockerised microservice containers and Kafka, with a (very) basic React frontend.

## Objectives

### Implement
- Backend comprising standalone microservices
- Kafka
- Redis
- Cassandra
- Protocol buffers
- Reporting of all viewable on a frontend admin panel/additional monitoring containers

### Then (most importantly)
- Create and destroy additional service containers, implement resiliency techniques until it keeps working.

## Frontend stuff
- Spool it up with `docker compose up -d`
- Go to http://localhost:5173
- Enter any user name, this will create a user (don't care about passwords here)
- Create an account (this will credit you with an initial balance because it's a very kind bank)
- Create more accounts by shifting your initial balance around
- Transfer between accounts

## What happens
When creating a new account, if it's the user's first account, the `account service` creates the account and create a message on the transactions topic to credit the account with a random amount.

If it's not the first account, it creates two transaction messages; one to credit the new account and one to debit the source account.

Transfers go to the `payment service`, if everything seems in order it will create messages on the `payment requested` topic. These will be picked up by the `account service` to do basic checks, such as does the source account exist and have the required funds. As a transfer between the user's accounts, it also verifies the source account is owned by the user. If all checks pass, the `payment service` issues messages for the `transaction service`.

## WIP stuff
- all of it really
- invalidate/reset Redis caches with a separate service that picks up messages relating to changed accounts
- switch from postgres to multiple sharded cassandra instances
- implement retry & dead letter topics
- add random delays to make things fail/time out/be racey
- make the front end show statuses of things when they're not instant
- send money to other users, validating some basic user info first as banks do
- add a DO LOTS OF THINGS admin mode button on the front end to initiate lots of account creations and payments to see what happens
