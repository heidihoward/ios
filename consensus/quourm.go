package consensus

import (
	"errors"
	"strconv"
	"strings"
)

// QuorumSys offers basic suport for various of quorum systems
// currently only "counting systems" are supported
type QuorumSys struct {
	Name            string
	RecoverySize    int
	ReplicationSize int
}

func NewQuorum(configName string, n int) (QuorumSys, error) {
	var qs QuorumSys
	qs.Name = "counting" // currently only "counting systems" are supported
	qType := strings.Split(configName, ":")
	// note that golang / will truncate
	switch qType[0] {
	case "strict majority":
		qs.ReplicationSize = (n / 2) + 1
		qs.RecoverySize = (n / 2) + 1
	case "non-strict majority":
		qs.ReplicationSize = (n + 1) / 2
		qs.RecoverySize = (n / 2) + 1
	case "all-in":
		qs.ReplicationSize = n
		qs.RecoverySize = 1
	case "one-in":
		qs.ReplicationSize = 1
		qs.RecoverySize = n
	case "fixed":
		i, err := strconv.Atoi(qType[1])
		if err != nil {
			return qs, errors.New("Quourm system is not recognised")
		}
		qs.ReplicationSize = i
		qs.RecoverySize = n + 1 - i
	default:
		return qs, errors.New("Quourm system is not recognised")
	}
	if qs.RecoverySize+qs.ReplicationSize <= n {
		return qs, errors.New("Unsafe quorum system has been chosen")
	}
	return qs, nil
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
	return count >= q.ReplicationSize
}

func (q QuorumSys) getReplicationQuourm(id int, n int) []int {
	quorum := make([]int, q.ReplicationSize)
	// TODO: consider replacing with random quorums
	for i := 0; i < q.ReplicationSize; i++ {
		quorum[i] = mod(i+id, n)
	}
	return quorum
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
