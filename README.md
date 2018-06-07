# White Water
White Water is a distributed Key Value Store built using the Raft algorithm in Golang.

## Project Structure
White Water is separated into 3 modules
1. `messages` which drives mesage sending and receiving
    - `messages.go` - Contains logic for ZeroMQ clients and ways to register message handlers and subscribe to specific types of messages.
    - `hello.go` - Contains a handler for hello messages from the ZeroMQ broker
2. `raft` which handles Raft consensus logic
    - `append.go` - Handles processing of Append Entries RPC messages
    - `get.go` - Handles processing of Get messages from a client.
    - `raft.go` - Contains core Raft logic with helper functions for message interaction and routing RPCs to respective handlers.
    - `requestVote.go` - Handles processing of Request Vote RPC messages and all other election process logic
    - `stuctures.go` - Holds all the data for the node structs and the RPC messages we use in Raft
    - `timer.go` - Contains utility functions for reseting the timer used in raft elections, timeouts, and heartbeats
3. `storage` which handles the state machine interface
    - `storage.go` - Contains structures for managing the state of the node and logs. Implements the get and set interfaces for updating the local model after a consensus round. Allows leaders to get local values from the model.

## Running White Water
To run White Water, first ensure you have Golang installed, we recommend Go 1.8+.

Next, ensure you have the ZeroMQ dependencies by installing `ZeroMQ version 4.0.1 or above. To use CURVE security in versions prior to 4.2, ZeroMQ must be installed with libsodium enabled.`

For additional instructions, see the [ZMQ4 Github Project](https://github.com/pebbe/zmq4).

Finally, run `go build` in our source repository.

Note that this requires pulling from GitHub and thus we have to make our project public in order for you to pull the dependencies for White Water.

Once you've built the source, you can stay in the directory that contains `node.go` and run `chistributed` as you would normally.

Our tests are inside `tests/` and can be run with the `--run` flag in chistributed.

For tests with 5 nodes, please include a `--config-file tests/five-nodes.conf` flag when running chistributed.

Example outputs of our test runs are provided in the `tests/` directory.

You'll note that our tests don't always look the same because leader election in Raft is nondeterministic, thus the leader node is not predictable when causing specific failures.

To get around this, we provided output from non-leader nodes that point to the current leader and also provide answers to get requests so tests are easier to run.


## Tests
Run tests from the impl directory:

1. Basic set and get behavior
chistributed --run scripts/basic-sets.chi

2. Multiple sets
chistributed --run scripts/multi-sets.chi

3. Numerous Sets and gets (> 80)
chistributed --run scripts/big-sets.chi

4. Basic Failures
chistributed --run scripts/simple-fail.chi

5. Multiple Different Failures
chistributed --run scripts/multi-fail.chi

6. Basis Partition
chistributed --run scripts/basic-partition.chi

7. Partition with Leader in majority
chistributed --config-file scripts/five-nodes.conf --run scripts/big-1-partition.chi

8. Partition with Leader in minority
chistributed --config-file scripts/five-nodes.conf --run scripts/big-2-partition.chi

Tests outputs from the creators of White Water, are in impl/scripts
and are of the form testname-sample.out

These tests were run on UChicago CS VM 201718.3 (Ubuntu 16.04.3 LTS)
