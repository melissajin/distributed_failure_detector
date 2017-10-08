## Setup
The script setup.sh takes configurations from setup.config.
The configuration file has these fields:
 - USAGE takes SETUP or STOP. 
    - SETUP prepares each VM by cloning or pulling from the git repo, setting up Go, and installing protobuf. 
    - STOP stops the servers and clears the ports that they were running on.
 - VM_NODES takes the host names of the VMs, separated by commas.
 - DIRECTORY is the directory holding the git repository.
 - GOPKG is the binary download of Go.
 - HOME is the home directory of the VMs.

To prepare the VMs, run `./setup.sh setup.config` with USAGE set to SETUP.

## Usage
1. ssh into a vm `$ ssh tkao4@fa17-cs425-g46-01.cs.illinois.edu`
2. Make sure you can run the go command `$ go`. If not, set the PATH enviorment variable `$  export PATH=$PATH:/usr/local/go/bin`
3. Run the server.
    ```
    $ cd CS425-MP2/src/
    $ go run server/main.go
    ```
4. If you get an error when you run the program, set the GOPATH enviorment variable `$ export GOPATH=/home/tkao4/CS425-MP2`
5. A command line will appear which recognizes four commands:
    * join - add the current machine to the system *
    * leave - remove the current machine from the system
    * list - list all machines in the current machine's membership list
    * id - print ID of current machine
*\*Note: we set machine 1, 2, 3, 4, and 5 (fa17-cs425-g46-0[1-5].cs.illinois.edu) as the entry machines. If any of these machine are not running, no other machines can join the system.*