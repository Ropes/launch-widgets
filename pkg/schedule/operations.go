package schedule

import (
	"strconv"

	"github.com/google/uuid"
	wyvv "github.com/ropes/whatsyourvectorvictor"
	log "github.com/sirupsen/logrus"
)

// Node apply costs
func NodesApplyCost(nodes []*Node) {
	for _, n := range nodes {
		n.InvoicePods()
	}
}

// Iterate workloads, if they have surpassed priority value, stop them,
// and return to be re-scheduled.
func ExpireLowPriorityPods(nodes []*Node, expirationPriority int) []*Pod {
	l := log.New()
	expiredPods := make([]*Pod, 0)
	for _, n := range nodes {
		for _, p := range n.Pods {
			if p.Priority > expirationPriority {
				l.WithFields(log.Fields{"pod": p, "priority": p.Priority, "node": n.Labels["name"]}).Info("stopping pod on node")
				err := n.Stop(p)
				if err != nil {
					l.WithFields(log.Fields{"pod": p, "node": n.Labels["name"]}).Error("error removing pod from node")
					continue
				}
				expiredPods = append(expiredPods, p)
			}
		}
	}
	return expiredPods
}

// NPods returns n number of pods with compute resources consuming
// 1 cpu and 1k bytes.
func NPods(n int) []*Pod {
	pods := make([]*Pod, 0)

	for i := 0; i < n; i++ {
		nstr := strconv.Itoa(i)
		p := &Pod{
			Labels: map[string]string{
				"name": nstr,
			},
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 1000,
				mem:  1000,
			},
			Priority: 0,
		}
		pods = append(pods, p)
	}
	return pods
}

// NNodes returns a set of nodes named from the NATO alphabet.
// Each node has 8CPU and 16k bytes memory.
func NNodes(n int) []*Node {
	nodes := make([]*Node, 0)
	for i := 0; i < n && i < len(wyvv.Alphabet); i++ {
		n := &Node{
			log: log.New(),
			Labels: map[string]string{
				"name": wyvv.Alphabet[i],
			},
			Pods: make(map[uuid.UUID]*Pod),
			Limit: ComputeResource{
				mCPU: 8000,
				mem:  16000,
			},
			PriorityCost: 10,
		}
		nodes = append(nodes, n)
	}
	return nodes
}
