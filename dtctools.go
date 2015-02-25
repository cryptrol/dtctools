package main

import (
	"log"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"net/http"
	"bytes"
	"flag"
	"errors"
	"io/ioutil"
	"path/filepath"
	"encoding/base64"
	"strings"
	"./envelope"
)

type Message struct {
	Jsonrpc string `json:"jsonrpc"`
	Id interface{} `json:"id,omitempty"`
	Method string `json:"method"`
	Params interface{} `json:"params"`
}

type Reply struct {
	Result interface{} `json:"result"`
	Error *Error `json:"error"`
	Id *interface{} `json:"id"`  // pointer so it can be null
}

type Error struct {
	Code int `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// Block structure
type Block struct {
	Primechain string `json:"primechain"`
	Previousblockhash string `json:"previousblockhash"`
	Nextblockhash string `json:"nextblockhash"`
	Hash string `json:"hash"`
	Confirmations int64 `json:"confirmations"`
	Tx []string `json:"tx"`
	Bits string `json:"bits"`
	Height int `json:"height"`
	Transition float64 `json:"transition"`
	Primeorigin string `json:"primeorigin"`
	Version int64 `json:"version"`
	Merkleroot string `json:"merkleroot"`
	Difficulty float64 `json:"difficulty"`
	Size int64 `json:"size"`
	Headerhash string `json:"headerhash"`
	Time float64 `json:"time"`
	Nonce int64 `json:"nonce"`
}

type Transaction struct {

}

func buildMessage(command string, args interface{}) ([]byte, error) {
	rawMessage := Message{"1.0", nil, command, args}
	finalMessage, err := json.Marshal(rawMessage)
	if err != nil {
		return nil, err
	}
	return finalMessage, nil
}

func getMessageReply(msg []byte) (Reply, error) {
	client := &http.Client{}
    var reply Reply
    resp, err := client.Post(daemonurl, apptype, bytes.NewReader(msg))
    if err != nil {
        return reply, errors.New("Error connecting to server." + err.Error())
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return reply,errors.New("HTTP status error: " + resp.Status)
    }
    decoder := json.NewDecoder(resp.Body)
    err = decoder.Decode(&reply)
    if err != nil {
        return reply,errors.New("Can't decode daemon reply :" + err.Error())
    }
    if reply.Error != nil {
    	log.Printf("Daemon returned error %v \n", reply.Error)
        return reply, errors.New("Daemon returned error.")
    }
    return reply, nil
}

const (
	apptype = "application/json"
)

var (
	daemonurl string
)

func main() {
	var user string
	flag.StringVar(&user, "user", "datacoinrpc", "The RPC user set in datacoin.conf")
	var password string
	flag.StringVar(&password, "password", "" , "The RPC password set in datacoin.conf")
	var server string
	flag.StringVar(&server, "server", "127.0.0.1", "The IP address of the RPC server")
	var port string
	flag.StringVar(&port, "port", "11777", "The port where the RPC server is listening to")
	var fromblock int
	flag.IntVar(&fromblock, "fromblock", 1, "The block where we will start dumping data from")
        var toblock int
        flag.IntVar(&toblock, "toblock", 720859, "The block where we will stop dumping data from")
	flag.Parse()
	
	daemonurl = "http://" + user + ":" + password+ "@" +server + ":" + port

	for i := fromblock; i <= toblock; i++ {
		// Get the block hash for block i
		var blocknum [1]int
		blocknum[0] = i
		msg, err := buildMessage("getblockhash", blocknum)
	    	if err != nil {
        		log.Fatal(err)
		}
		var r Reply
		r, err = getMessageReply(msg)
        	if err != nil {
	            log.Fatal(err)
        	}
		var param[1]string
        	param[0] = r.Result.(string)
		// Let's request the block
		msg, err = buildMessage("getblock", param)
		if err != nil {
			log.Fatal(err)
		}
		r, err = getMessageReply(msg)
		if err != nil {
			log.Fatal(err)
		}
		// We have the block, check every transaction for data
		var cf map[string]interface{}
		cf = r.Result.(map[string]interface{})
		var ta []interface{}
		ta = cf["tx"].([]interface{})
		for j := 0 ; j < len(ta) ; j++ {
			// Prepare RPC message
			msg, err = buildMessage("getdata", []string{ta[j].(string)})
			if err != nil {
				log.Fatal(err)
			}
			// Send RPC message and get reply
			r, err = getMessageReply(msg)
			if err != nil {
				log.Fatal(err)
			}
			rd := r.Result.(string)
			// Check wether transaction data field is empty
			if rd != "" {
				filecontents, err := base64.StdEncoding.DecodeString(rd)
				if err != nil {
					log.Fatal("Error decoding data." + err.Error())
				}
				// Try to Unmarshal the object as a protocol buffer message.
				env := &envelope.Envelope{}
				err = proto.Unmarshal(filecontents, env)
				if err != nil {
					log.Printf("Found raw data on block %s for transaction %s ",i , ta[j].(string))
					// Since it's not an Envelope type message, let's get the raw data and save to file
					mimetype := http.DetectContentType(filecontents) // type guessing
					mimetype = strings.Replace(mimetype, "/", "-", -1) 
		    			err = ioutil.WriteFile(ta[j].(string) + "." + mimetype, filecontents, 0644)
		    			if err != nil {
	    					log.Fatal("Can't write to file.\n" + err.Error())
	    				}
				} else {
					// TODO check multipart files
					log.Printf("Found envelope data on block %s for transaction %s ",i , ta[j].(string))
					var filename string
					switch env.GetCompression() {
						case 0 : filename = env.GetFileName()
						case 1 : filename = env.GetFileName() + ".bzip2"	
						case 3 : filename = env.GetFileName() + ".xz"
					}
					// Some filenames may contain the full path so keep just the base
					filename = filepath.Base(filename)
					err = ioutil.WriteFile(filename, env.GetData(), 0644)
		    			if err != nil {
		    				log.Fatal("Error writing envelope data to file. " + err.Error())
	    				}
				}
			}
		}

	} // endfor
} // endmain

