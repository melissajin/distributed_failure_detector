#!/bin/bash
 
############################## PRE PROCESSING ################################
#check and process arguments
REQUIRED_NUMBER_OF_ARGUMENTS=1
if [ $# -lt $REQUIRED_NUMBER_OF_ARGUMENTS ]
then
    echo "Usage: $0 <path_to_config_file>"
    exit 1
fi

CONFIG_FILE=$1
 
echo "Config file is $CONFIG_FILE"
echo ""
 
#get the configuration parameters
source $CONFIG_FILE

############################## SETUP ################################
if [ "$USAGE" == "SETUP" ]
then
	counter=1
	for node in ${VM_NODES//,/ }
	do
		echo "Setting up $node ..."
	    COMMAND=''

	    # Get code from git repository.
	    COMMAND=$COMMAND"
	    if [ ! -d \"$DIRECTORY\" ]; then
	    	echo \"Cloning repository...\";
	    	git clone https://tkao4:Ilmjsm2696@gitlab.engr.illinois.edu/tkao4/CS425-MP2.git;
	    else
	    	echo \"Pulling from repository...\";
	    	cd CS425-MP2/src;
	    	git pull https://tkao4:Ilmjsm2696@gitlab.engr.illinois.edu/tkao4/CS425-MP2.git;
	    fi; "

	    # Install Go
	    COMMAND=$COMMAND"
	    if [ ! -e \"$GOPKG\" ]; then
	    	echo \"Installing Go...\";
		    wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz;
		    sudo tar -C /usr/local -xvzf go1.7.3.linux-amd64.tar.gz;
		fi;"

		# Install Protobuf
		COMMAND=$COMMAND"
	    if [ ! -e \"$PROTOBUF_DIR\" ]; then
	    	echo \"Installing protobuf...\";
	    	rm -rf github.com/;
	    	go get -u github.com/golang/protobuf/protoc-gen-go;
		    export GOPATH=/home/tkao4/CS425-MP2;
		    export PATH=$PATH:$GOPATH/bin;
		fi;"

	    let counter=counter+1 
	    ssh -t -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no $node "
	            $COMMAND"
	done
fi


