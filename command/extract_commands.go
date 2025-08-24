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

func ExtractCommandsFromParsedData(parsedData []string) ([]Cmd, error) {
	commands := make([]Cmd, 0)

	// The for loop adds 1 to i for each command, so if the command takes 2 arguments, 
	// i will be incremented by 2, then the next command will start at i+2 (added by us 1) + 1 (added by the for loop) 
	// so the next command will start at i+3 to skip the arguments of the previous command
	for i := 0; i < len(parsedData); i++ {
		switch strings.ToLower(parsedData[i]) {
		case protocol.PING:
			commands = append(commands, Cmd{Name: protocol.PING})
		case protocol.ECHO:
			if i+1 >= len(parsedData) {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'echo' command")
			}
			commands = append(commands, Cmd{Name: protocol.ECHO, Args: []string{parsedData[i+1]}})
			i++
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

			args := []string{parsedData[i+1], parsedData[i+2]}

			// check PX and EX
			for j := i + 3; j < len(parsedData); j++ {
				if strings.ToLower(parsedData[j]) == protocol.PX || strings.ToLower(parsedData[j]) == protocol.EX {
					args = append(args, parsedData[j], parsedData[j+1])
					j++
				}
			}

			commands = append(commands, Cmd{Name: protocol.SET, Args: args})
			i += len(args) - 1

		case protocol.RPUSH:
			args := parsedData[i+1:]
			commands = append(commands, Cmd{Name: protocol.RPUSH, Args: args})
			i = len(parsedData) - 1

		case protocol.LPUSH:
			args := parsedData[i+1:]
			commands = append(commands, Cmd{Name: protocol.LPUSH, Args: args})
			i = len(parsedData) - 1

		case protocol.LRANGE:
			args := parsedData[i+1:]
			commands = append(commands, Cmd{Name: protocol.LRANGE, Args: args})
			i = len(parsedData) - 1

		case protocol.LLEN:
			commands = append(commands, Cmd{Name: protocol.LLEN, Args: []string{parsedData[i+1]}})
			i++

		case protocol.LPOP:
			args := []string{parsedData[i+1]}

			if checkArrayLen(i, len(parsedData), 2) {
				if isNumber(parsedData[i+2]){
					args = append(args, parsedData[i+2])
					i+=2
				}else{
					i++
				}
			}
			commands = append(commands, Cmd{Name: protocol.LPOP, Args: args})

		case protocol.BLPOP:
			args := []string{parsedData[i+1]}

			if checkArrayLen(i, len(parsedData), 2) {
				if isNumber(parsedData[i+2]){
					args = append(args, parsedData[i+2])
					i+=2
				}else{
					i++
				}
			}
			commands = append(commands, Cmd{Name: protocol.BLPOP, Args: args})

		case protocol.TYPE:
			if !checkArrayLen(i, len(parsedData), 1){
				continue
			}

			commands = append(commands, Cmd{Name: protocol.TYPE, Args: []string{parsedData[i+1]}})
			i++

		case protocol.XADD:
			args := parsedData[i+1:]
			commands = append(commands, Cmd{Name: protocol.XADD, Args: args})
			i = len(parsedData) - 1

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
			commands = append(commands, Cmd{Name: protocol.PSYNC, Args: []string{parsedData[i+1]}})
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

func checkArrayLen(currentIndex, arrayLen, requiredArgs int) bool {
	if currentIndex + requiredArgs >= arrayLen {
		return false
	}
	return true
}

func isNumber(s string) bool{
	if len(s) == 0 {
		return false
	}

	for _,c := range s {

		if c == '.'{
			continue
		}

		if c < '0' || c > '9' {
            return false
        }
	}
	return true
}
