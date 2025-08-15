package command

import (
	"fmt"
	protocol "redisgo/protocol"
	"strings"
)

type Cmd struct {
	Name string
	Args []string
}

func extractCommandsFromParsedData(parsedData []string) ([]Cmd, error) {
	commands := make([]Cmd, 0)
	for i := 0; i < len(parsedData); i++ {
		switch strings.ToLower(parsedData[i]) {
		case protocol.PING:
			commands = append(commands, Cmd{Name: protocol.PING})
		case protocol.ECHO:
			if i+1 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'echo' command")
			}
			commands = append(commands, Cmd{Name: protocol.ECHO, Args: []string{parsedData[i+1]}})
		case protocol.GET:
			if i+1 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'get' command")
			}
			commands = append(commands, Cmd{Name: protocol.GET, Args: []string{parsedData[i+1]}})
			i++
		case protocol.SET:
			if i+2 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'set' command")
			}

			args := []string{protocol.SET, parsedData[i+1], parsedData[i+2]}

			// check PX and EX
			for j := i + 3; j < len(parsedData); j++ {
				if strings.ToLower(parsedData[j]) == protocol.PX || strings.ToLower(parsedData[j]) == protocol.EX {
					args = append(args, parsedData[j], parsedData[j+1])
					j++
				}
			}

			commands = append(commands, Cmd{Name: protocol.SET, Args: args})
			i += len(args) - 1
		case protocol.INFO:
			if i+1 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'info' command")
			}
			commands = append(commands, Cmd{Name: protocol.INFO, Args: []string{parsedData[i+1]}})
			i++
		case protocol.REPLCONF:
			if i+2 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'replconf' command")
			}
			commands = append(commands, Cmd{Name: protocol.REPLCONF, Args: []string{parsedData[i+1], parsedData[i+2]}})
			i += 2
		case protocol.FULLRESYNC:
			if i+2 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'fullresync' command")
			}
			commands = append(commands, Cmd{Name: protocol.FULLRESYNC, Args: []string{parsedData[i+1], parsedData[i+2]}})
			i += 2
		case protocol.PSYNC:
			if i+1 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'psync' command")
			}
			commands = append(commands, Cmd{Name: protocol.PSYNC,Args: []string{parsedData[i+1]}})
			i++
		case protocol.WAIT:
			if i+2 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'wait' command")
			}
			commands = append(commands, Cmd{Name: protocol.WAIT, Args: []string{parsedData[i+1], parsedData[i+2]}})
			i += 2
		}
	}
	return commands, nil
}
