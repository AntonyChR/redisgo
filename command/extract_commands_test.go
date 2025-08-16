package command

import (
	"redisgo/protocol"
	"testing"
)

type TestCase struct {
	input    []string
	expected Cmd
}

func TestExtractCommands(t *testing.T) {
	cases := []TestCase{
		{
			input:    []string{"SET", "foo", "bar"},
			expected: Cmd{protocol.SET, []string{"foo", "bar"}},
		},
		{
			input:    []string{"GET", "foo"},
			expected: Cmd{protocol.GET, []string{"foo"}},
		},
		{
			input:    []string{"PING"},
			expected: Cmd{Name: protocol.PING},
		},
		{input: []string{"ECHO", "Hello"},
			expected: Cmd{protocol.ECHO, []string{"Hello"}},
		},
		{input: []string{"INFO", "memory"},
			expected: Cmd{protocol.INFO, []string{"memory"}},
		},
	}

	for i, c := range cases {
		commands, err := ExtractCommandsFromParsedData(c.input)
		if err != nil {
			t.Errorf("case [%d]: %v", i, err)
		}

		if len(commands) != 1 {
			t.Errorf("case [%d]: expected 1 command, got %d", i, len(commands))
		}

		cmd := commands[0]

		if c.expected.Name != cmd.Name {
			t.Errorf("case [%d]: invalid name, expected \"%s\", got \"%s\"", i, c.expected.Name, cmd.Name)
		}

		if len(c.expected.Args) != len(cmd.Args) {
			t.Errorf("case [%d]: inconsistent number of arguments, expected \"%d\", got \"%d\"", i, len(c.expected.Args), len(cmd.Args))
		}

		for j, expectedArg := range c.expected.Args {
			if expectedArg != cmd.Args[j] {
				t.Errorf("case [%d]: invalid argument at index %d, expected \"%s\", got \"%s\"", i, j, expectedArg, cmd.Args[j])
			}
		}

	}
}
