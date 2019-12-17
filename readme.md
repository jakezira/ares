# Ares

**Ares** is the monitoring system, which provides level 1 high availability.

## Introduction

- Ares manager

  There's single **Ares manager**, it communicates with **Ares node**s, retrieves the application status, and organizes the application running.

- Ares node

  It runs on every computer, monitors bunches of applications, reports application status to **Ares manager**, and when there's new command from **Ares manager**, it runs again.

## Configuration

- Ares manager

  ```json
  {
    "port": "4261",
    "monitor": "5261",
    "debug": "on"
  }
  ```

  Ares manager runs the grpc service on `port`.

- Ares node

  - `name` is the node name.

  - `manager` is the `Ares manager` address.

  - `apps` is about the applications current `Ares node` supports.

    - `name` is the app name.
    - `run` is the application relative path.
    - `dir` is the working directory to run the app.

  ```json
  {
    "port": "4262",
    "monitor": "5262",

    "name": "node1",
    "manager": {
        "host": "10.70.3.98",
        "port": "4261"
    },
    "apps": [
        {
            "name": "app1",
            "run": ".\\App1.exe",
            "dir": "E:\\App1\\_release"
        },
        {
            "name": "app2",
            "run": ".\\App2.exe",
            "dir": "E:\\App2\\_release"
        }
    ]
  }
  ```

## Run

- Run `Ares manager` on server computer.
- Run `Ares node` on each computer, with the configuration set.
  - `Ares node` connects to `Ares manager` and registers applications information it supports.
- Run `la` on `Ares manager` to list applications the system has.
- Run `ln` on `Ares manager` to list nodes the system has.
- Run `on {appName}` on `Ares manager` to run the application on one of node pc.
- When the application started by the Ares system crashes, it makes another run automatically.
