# WB level 0
My test task at Wildberries
> Develop a demo service with a simple interface that displays order data. The data model in
> the JSON format is attached to the task.

## Content
- [Task](#task)
- [Third-party modules](#modules)
- [Usage](#usage)
- [To do](#todo)


## Task
1. Deploy PostgreSQL locally
2. Develop a service
    - Implement channel connection and subscription using nats-streaming
    - Write the received data to the database
    - Implement caching of the received data
    - In case of a service crash, restore the cache from the database 
    - Start the http server and output data by id from the cache
3. develop simple interface for displaying the received data by order id

## Modules
- [NATS Streaming](https://github.com/nats-io/stan.go)
- [pgx](https://github.com/jackc/pgx)
- [Chi](https://github.com/go-chi/chi)
- [cleanenv](https://github.com/ilyakaznacheev/cleanenv)
- [godotenv](https://github.com/joho/godotenv)

## Deploy
1. Install golang on your machine
2. Clone repo:
```
git clone https://github.com/asoloshchenko/wb_l0.git
```
3. Fill in the .yaml file with parameters. Additionally, you need to set the environment variable CONFIG_PATH to the path of your .yaml file (or fill .env file).
4. Run service:
```
    task run
```
5. Run publisher:
```
    task pub
```
## Usage
After the server starts, you can send a GET request to `yoururl/api/{id}` where `{id}` is the ID of your order. If the order is stored in the cache or the database, the response body will consist of the full JSON of the order. If you also run a publisher, you can send fake data to NATS streaming. If the message is stored successfully, you'll get the ID of the sent message back.


## TODO
- [ ] Сover with tests
- [ ] Wrap in a docker
- [ ] Add normal deploy
- [ ] ...

