package negotiation

import (
	"fmt"
	"log"
	"net"
	"redisgo/protocol"
)

type ReplicaController struct {
	replicas []Replica
	parser   protocol.Parser
}

// TODO: remove hardcoded content
var BASE64_EMPTY_RDB_FILE = "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="

func (rc *ReplicaController) Setup(port string, host string) error {
	conn, err := net.Dial("tcp", host+":"+port)
	buff := make([]byte, 1024)
	if err != nil {
		return err
	}

	// send PING
	pingCmd := rc.parser.EncodeAsArray([]string{"PING"})
	if _, err = conn.Write([]byte(pingCmd)); err != nil {
		return err
	}

	// wait for PONG
	n, err := conn.Read(buff)
	if err != nil {
		return err
	}

	if parsedResp, err := rc.parser.Decode(buff[:n]); err != nil {
		return err
	} else if string(parsedResp[0]) != "PONG" {
		return fmt.Errorf("unexpected response the master, got \"%v\", expected \"PONG\"\n", parsedResp)
	}

	// send port with REPLCONF listening-port command
	// TODO: replace harcoded port value
	replConfCmd := rc.parser.EncodeAsArray([]string{"REPLCONF", "listening-port", "3001"})
	if _, err = conn.Write([]byte(replConfCmd)); err != nil {
		return nil
	}

	// wait "OK"

	n, err = conn.Read(buff)
	if err != nil {
		return err
	}
	if parsedResp, err := rc.parser.Decode(buff[:n]); err != nil {
		return err
	} else if string(parsedResp[0]) != "OK" {
		return fmt.Errorf("unexpected response the master, got \"%v\", expected \"PONG\"\n", parsedResp)
	}

	// send capability sync2

	replConfCmd = rc.parser.EncodeAsArray([]string{"REPLCONF", "capa", "sync2"})
	if _, err = conn.Write([]byte(replConfCmd)); err != nil {
		return err
	}

	// wait "OK"
	n, err = conn.Read(buff)
	if err != nil {
		return err
	}
	if parsedResp, err := rc.parser.Decode(buff[:n]); err != nil {
		return err
	} else if string(parsedResp[0]) != "OK" {
		return fmt.Errorf("unexpected response the master, got \"%v\", expected \"PONG\"", parsedResp)
	}

	// Send PSYNC ? -1
	pSyncCmd := rc.parser.EncodeAsArray([]string{"PSYNC", "?", "-1"})
	if _, err = conn.Write([]byte(pSyncCmd)); err != nil {
		return err
	}

	// wait FULLRESYNC
	n, err = conn.Read(buff)
	if err != nil {
		return err
	}

	fullReSync, err := rc.parser.Decode(buff[:n])
	if err != nil {
		return err
	} else if len(fullReSync) != 3 {
		return fmt.Errorf("unexpected number of parameters for the FULLRESYNC command, got %d, expected %d", len(fullReSync), 3)
	}

	// TODO: set master info
	//masterReplId := fullReSync[0]
	//masterReplOffset := fullReSync[1]

	// wait RDB file

	n, err = conn.Read(buff)
	if err != nil {
		return err
	}

	// TODO: decode RDB file content
	if rdbFileContent, err := rc.parser.Decode(buff[:n]); err != nil {
		return err
	} else if rdbFileContent[0] != BASE64_EMPTY_RDB_FILE {
		return fmt.Errorf("unexpected data, got \"%s\", expected \"%s\"", rdbFileContent[0], BASE64_EMPTY_RDB_FILE)
	}

	log.Println("[ReplicaController] Replica setup completed successfully")

	return nil

}

type Replica struct {
}
