# remitly-api


## About
This is a simple CRUD for fast retrieval of SWIFT data using Redis + GO.

## How to run
Before installation make sure you have Docker installed

1. Copy the repo:

```bash
git clone https://github.com/grysj/remitly-api.git
```

2. In a root directory of the cloned repo, run:

```bash
docker compose build
```

3. To start the container, run (while still in the root directory):
```bash
docker compose up
```

## Usage
There already is a `SWIFT_CODES.csv` in the repo. If you want to change something,  eg. the csv file, simply swap out path to the file in `.env`
```bash
# .env
CV_PATH="<pathToYourCSV>"
```


## How to test
In a root directory, run
```bash
docker compose run test
```


