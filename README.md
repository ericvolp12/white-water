# White Water
White Water is a distributed Key Value Store built using the Raft algorithm in Golang.

## Project Structure
White Water is separated into 4 modules
1. `messages` which drives mesage sending and receiving
2. `raft` which handles Raft consensus logic
3. `storage` which handles the state machine interface

