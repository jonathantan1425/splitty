# Splitty - Telegram Bot to Split Bills

A Go-based Telegram Bot to manage and track bill splitting among users within the same Telegram group.

## Project Overview

Traveling takes a huge part during our exchange period here in St. Gallen, and we often go traveling with friends. Inevitably, to pay for entry tickets or bills, one person often pays the bill first and the group split it later, which emphasizes the importance of having a useful tool to help us split the bill.

Even though there are already applications such as Splitwise or Settleup, we feel like it will be more convenient if we can do it through communication applications, which is the main channel for us to communicate and keep contact with each other. Since in​​ Telegram, users can induce chatbots to complete certain tasks, such as Bus Uncle that help users get control of the schedules of buses, we decided to create a chatbot that helps users to split, settle, and balance the bills automatically. Aftering inducing it into Telegram, users can split the bills in their group chats and therefore clarify the bills without any confusion. 

## Prerequisites

- Recommended platform - Linux (Ubuntu 20.04)
- Docker installed

### Verify Docker
Run `docker` in terminal to check if Docker is installed. An example of a positive response:
```bash
$ docker
Usage:  docker [OPTIONS] COMMAND

A self-sufficient runtime for containers

Options:
...
```
If `Command 'docker' not found...`, proceed with installing Docker
### Install Docker
Refer to https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-20-04 for instructions.

## Installation and set-up

The Telegram Bot can be executed by (1) setting up local directories, and (2) providing the Telegram API Key (this can be retrieved on request from the authors)

### Set-up local directories

Clone this repository or download the files to a local directory.\
Open a terminal session and navigate to the path of this repository/codebase.
> e.g. if working path is `/usr/lib/splitty`
```bash
cd /usr/lib/splitty
```

### Provide the Telegram API Key

Interacting on the Telegram platfrom requires a Telegram API Key. For more information, refer to: https://core.telegram.org/bots/tutorial

1. Provide the Telegram API Key in `.env.example` using any text editor
2. Rename `.env.example` to `.env`.

__Note: `.env` is automatically ignored by git__
```yaml
TELEGRAM_API_KEY=<insert key here>
```

### Run the Bot as a Docker Image
A docker compose configuration has been set up for ease of execution. If docker compose is unavailable, fall back to `docker build` and `docker run`

```bash
docker compose up --build -d
```

### Successful Set-up
On successful start up, you should see the following logs

```bash
Attaching to splitty
splitty  | XXXX/XX/XX XX:XX:XX INFO Opening database
splitty  | XXXX/XX/XX XX:XX:XX INFO Database init complete
```

### Troubleshooting

If you encounter this error: `splitty  | XXXX/XX/XX XX:XX:XX ERROR Error creating table: !BADKEY="unable to open database file: no such file or directory"`, it is likely due to a missing `./data/sqlite.db` file. Create the file and it should resolve the issue.

```bash
mkdir data
touch ./data/sqlite.db
```

## Navigating the Telegram Bot

Use the following commands to access its features

### `/help`
Provides a list of commands that it will accept.

### `/split`
Initiates a splitting action. Reply the message with the following format: `## @user1 @user2` where `##` refers to the numerical amount to be split and `@user1 @user2` refers to the other users to be involved in the splitting.

Splits are done on an equal basis currently and by default includes the author of the message as a participant. E.g. tagging 2 other members will split the amount three-ways (2 + author)

### `/balance`
Retrieves and displays the amount owed in the group chat.

### `/settle`
Initiates a settle action. Reply the message with the following format: `## @user1` where `##` refers to the numerical amount to be settled and `@user1` refers to the other user that is being settled with.

Currently, only 1 user can be settled with at a time.

## Actual Telegram Screen Display of the Bot

Using the `/split` function allows you to evenly split the amount paid for the other person and record the transaction for all parties.

<img src="/assets/split.png">
<img src="/assets/split_2.png">

Using the `/balance` command allows you to check the current balance of mutual debts between everyone.

<img src="/assets/balance.png">

Using the`/settle` command serves as a repayment function. If you have already repaid the other person, you can use this command to clear the debt.

Then you can see that after using `/settle`, the amount shown in `/balance` reflects the new amount.

<img src="/assets/settle.png">

## Acknowledgements

Project was built for the University of St.Gallen's course **8,789,1.00: Skills: Programming with Advanced Computer Languages**.

Contributed by:
- Jonathan Wui Heong Tan (23-627-102)
- Wei-Zhen Lee (23-626-815)
- Yi-Chin Huang (23-626-724)
