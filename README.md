# dtctools

This tool connects to a local Datacoin daemon and uses RPC commands to dump the data extracted from the transactions into files. It supports :
* Envelope protol buffers format as defined by the original Datacoin author. This will create a file with the original filename.
* Raw data with a wild guess on the data type. The tool will output the data to a filename consisting in the TX id and an extension based on the MIME type of the file (a guess).
Data is dumped to the current working dir.

# Datacoin configuration

You must set the server variable to 1 in order to allow RPC connections (by default only localhost will be able to connect).
Example contents of the `datacoin.conf` file :

    rpcuser=a_username
    rpcpassword=a_random_password
    server=1

# Usage

In order to build this tool you will need a working GO environment, and then :

    git clone https://github.com/cryptrol/dtctools
    cd dtctools
    go get github.com/golang/protobuf/proto
    go install

For help using dtctools type :

    dtctools --help

Provided the above example datacoin.conf file is used, this command would dump the data found in blocks 7000000 to 720000 (provided there is any) :

    dtctools -user=a_username -password=a_random_password -fromblock=700000 -toblock=720000

