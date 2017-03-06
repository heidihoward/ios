package consensus

import (
	"github.com/golang/glog"
	"strconv"
	"strings"
)

// QuorumSys offers basic suport for various of quorum systems
// currently only "counting systems" are supported
type QuorumSys struct {
	Name          string
	RecoverySize  int
	ReplicateSize int
}

func NewQuorum(configName string, n int) QuorumSys {
	var replication int
	var recovery int
	qType := strings.Split(configName, ":")
	// note that golang / will truncate
	switch qType[0] {
	case "strict majority":
		replication = (n / 2) + 1
		recovery = (n / 2) + 1
	case "non-strict majority":
		replication = (n + 1) / 2
		recovery = (n / 2) + 1
	case "all-in":
		replication = n
		recovery = 1
	case "one-in":
		replication = 1
		recovery = n
	case "fixed":
		i, err := strconv.Atoi(qType[1])
		if err != nil {
			glog.Fatal("Quourm system is not recognised")
		}
		replication = i
		recovery = n + 1 - replication
	default:
		glog.Fatal("Quourm system is not recognised")
	}
	if recovery+replication <= n {
		glog.Fatal("Unsafe quorum system has been chosen")
	}
	return QuorumSys{"counting", recovery, replication}
}

func (q QuorumSys) checkReplicationQuorum(nodes []bool) bool {
	// count responses
	count := 0
	for _, node := range nodes {
		if node {
			count++
		}
	}
	// check if responses are sufficient
	return count >= q.ReplicateSize
}

func (q QuorumSys) checkRecoveryQuorum(nodes []bool) bool {
	// count responses
	count := 0
	for _, node := range nodes {
		if node {
			count++
		}
	}
	// check if responses are sufficient
	return count >= q.RecoverySize
}
