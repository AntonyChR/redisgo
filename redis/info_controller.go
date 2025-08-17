package redis

import (
	"strconv"
	"strings"
)

type InfoController struct {
	Host                       string
	Port                       string
	Role                       string
	ConnectedSlaves            string
	MasterFailoverState        string
	MasterReplid               string
	MasterReplOffset           int
	SecondReplOffset           int
	ReplBacklogActive          string
	ReplBacklogSize            string
	ReplBacklogFirstByteOffset string
	ReplBacklogHistlen         string
}

const (
	ROLE                           = "role"
	PORT                           = "port"
	CONNECTED_SLAVES               = "connected_slaves"
	MASTER_FAILOVER_STATE          = "master_failover_state"
	MASTER_REPLID                  = "master_replid"
	MASTER_REPL_OFFSET             = "master_repl_offset"
	SECOND_REPL_OFFSET             = "second_repl_offset"
	REPL_BACKLOG_ACTIVE            = "repl_backlog_active"
	REPL_BACKLOG_SIZE              = "repl_backlog_size"
	REPL_BACKLOG_FIRST_BYTE_OFFSET = "repl_backlog_first_byte_offset"
	REPL_BACKLOG_HISTLEN           = "repl_backlog_histlen"
)

func (i *InfoController) GetFormattedInfo() string {
	var builder strings.Builder

	if i.Role != "" {
		builder.WriteString(ROLE + ":" + i.Role + "\r\n")
	}

	if i.Host != "" {
		builder.WriteString(ROLE + ":" + i.Role + "\r\n")
	}
	if i.Port != "" {
		builder.WriteString(PORT + ":" + i.Port + "\r\n")
	}
	if i.ConnectedSlaves != "" {
		builder.WriteString(CONNECTED_SLAVES + ":" + i.ConnectedSlaves + "\r\n")
	}
	if i.MasterFailoverState != "" {
		builder.WriteString(MASTER_FAILOVER_STATE + ":" + i.MasterFailoverState + "\r\n")
	}
	if i.MasterReplid != "" {
		builder.WriteString(MASTER_REPLID + ":" + i.MasterReplid + "\r\n")
	}
	builder.WriteString(MASTER_REPL_OFFSET + ":" + strconv.Itoa(i.MasterReplOffset) + "\r\n")

	if i.SecondReplOffset != 0 {
		builder.WriteString(SECOND_REPL_OFFSET + ":" + strconv.Itoa(i.SecondReplOffset) + "\r\n")
	}
	if i.ReplBacklogActive != "" {
		builder.WriteString(REPL_BACKLOG_ACTIVE + ":" + i.ReplBacklogActive + "\r\n")
	}
	if i.ReplBacklogSize != "" {
		builder.WriteString(REPL_BACKLOG_SIZE + ":" + i.ReplBacklogSize + "\r\n")
	}
	if i.ReplBacklogFirstByteOffset != "" {
		builder.WriteString(REPL_BACKLOG_FIRST_BYTE_OFFSET + ":" + i.ReplBacklogFirstByteOffset + "\r\n")
	}
	if i.ReplBacklogHistlen != "" {
		builder.WriteString(REPL_BACKLOG_HISTLEN + ":" + i.ReplBacklogHistlen + "\r\n")
	}

	return builder.String()
}
