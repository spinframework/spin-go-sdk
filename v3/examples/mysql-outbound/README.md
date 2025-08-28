# Requirements
- Latest version of [TinyGo](https://tinygo.org/getting-started/)
- Latest version of [Docker](https://docs.docker.com/get-started/get-docker/)

# Usage

In a terminal window, use the below command to run MySQL:
```sh
docker compose up -d
```

Then, you'll build and run your Spin app:
```sh
spin up --build
```

In another terminal window, you can interact with the Spin app:
```sh
curl localhost:3000
```

You should see the output:
```json
[{"ID":1,"Name":"Splodge","Prey":null,"IsFinicky":false},{"ID":2,"Name":"Kiki","Prey":"Cicadas","IsFinicky":false},{"ID":3,"Name":"Slats","Prey":"Temptations","IsFinicky":true},{"ID":4,"Name":"Maya","Prey":"bananas","IsFinicky":true},{"ID":5,"Name":"Copper","Prey":"Foxes","IsFinicky":false}]
```

To stop and clean up the MySQL container, run the following:
```sh
docker compose down -v
```