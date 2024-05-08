package main

import (
    "log"

    "github.com/joho/godotenv"
)

func Env_init() {

    err := godotenv.Load(".env")

    if err != nil {
        log.Fatal("Error loading .env file")
    }
}