package schedule

import (
	"os"
	"testing"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func pods(basePriority int) []*Pod {
	return []*Pod{
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 100,
				mem:  100,
			},
			Priority: 0 + basePriority,
		},
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 100,
				mem:  100,
			},
			Priority: 1000 + basePriority, // Extreme low priority
		},
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 500,
				mem:  1000,
			},
			Priority: 0 + basePriority,
		},
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 1000,
				mem:  1000,
			},
			Priority: 0 + basePriority,
		},
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 1000,
				mem:  1000,
			},
			Priority: 0 + basePriority,
		},
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 100,
				mem:  100,
			},
			Priority: 0 + basePriority,
		},
		{
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 1000,
				mem:  1000,
			},
			Priority: 0 + basePriority,
		},
		{ // Massive pod, will not be scheduled
			ID: uuid.New(),
			CR: ComputeResource{
				mCPU: 10000,
				mem:  10000,
			},
			Priority: 0 + basePriority,
		},
	}
}

func nodes() []*Node {
	return []*Node{
		{
			Pods: make(map[uuid.UUID]*Pod),
			Limit: ComputeResource{
				mCPU: 4000,
				mem:  16000,
			},
			PriorityCost: 10,
			log:          log.New(),
		},
		{
			Pods: make(map[uuid.UUID]*Pod),
			Limit: ComputeResource{
				mCPU: 200,
				mem:  16000,
			},
			PriorityCost: 10,
			log:          log.New(),
		},
		{
			Pods: make(map[uuid.UUID]*Pod),
			Limit: ComputeResource{
				mCPU: 16000,
				mem:  16000,
			},
			PriorityCost: 10,
			log:          log.New(),
		},
	}
}

func TestSchedulePodBasics(t *testing.T) {
	pList := pods(0)[:1]
	nList := nodes()[:1]
	// Test with only one pod and one node

	s := NewScheduler(pList, -25)
	s.ScheduleNodes(nList)

	node := nList[0]
	pod := pList[0]

	assert.Len(t, node.Pods, 1)
	node.Stop(pod)
	assert.Len(t, node.Pods, 0)
}

func TestMultiPodScheduling(t *testing.T) {
	pList := pods(0)
	nList := nodes()[:1]
	pLen0 := len(pList)

	s := NewScheduler(pList, -100)
	s.ScheduleNodes(nList)
	pLen1 := len(s.podsToStart)
	assert.NotEqual(t, pLen0, pLen1)
	t.Logf("Scheduler only has %d pods to start after %d", pLen1, pLen0)

	t.Run("test un-scheduling after priority changes", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			NodesApplyCost(nList)
		}

		expiredPods := ExpireLowPriorityPods(nList, 10)
		assert.Greater(t, len(expiredPods), 3)

		s.AddPodsToSchedule(expiredPods)
		pLen2 := len(s.podsToStart)
		assert.Equal(t, pLen0, pLen2)
		assert.Len(t, nList[0].Pods, 0)

		t.Run("Reset Pod Priorities", func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				s.AdjustPodPriority()
			}
			for _, p := range s.sortPods() {
				assert.Equal(t, p.Priority, 0, "Assert that the Scheduler is set to start all Pods again.")
			}
		})
	})
}

func TestMultiPodMultiNodeScheduling(t *testing.T) {
	pList := pods(0)
	nList := nodes()
	pLen0 := len(pList)

	s := NewScheduler(pList, -100)
	s.ScheduleNodes(nList)
	pLen1 := len(s.podsToStart)
	assert.NotEqual(t, pLen0, pLen1)
	t.Logf("Scheduler only has %d pods to start after %d", pLen1, pLen0)

	t.Run("test un-scheduling after priority changes", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			NodesApplyCost(nList)
		}

		expiredPods := ExpireLowPriorityPods(nList, 10)
		assert.Greater(t, len(expiredPods), 3)

		s.AddPodsToSchedule(expiredPods)
		pLen2 := len(s.podsToStart)
		assert.Equal(t, pLen0, pLen2)
		assert.Len(t, nList[0].Pods, 0)

		t.Run("Reset Pod Priorities", func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				s.AdjustPodPriority()
			}
			for _, p := range s.sortPods() {
				assert.Equal(t, p.Priority, 0, "Assert that the Scheduler is set to start all Pods again.")
			}
		})
	})
}

func TestEvenScheduling(t *testing.T) {
	nodeCount := 4
	podCountPerNode := 3
	nodes := NNodes(nodeCount)
	pods := NPods(nodeCount * podCountPerNode)
	t.Logf("Scheduling %d pods to %d nodes", len(pods), len(nodes))

	s := NewScheduler(pods, -100)
	s.ScheduleNodes(nodes)
	s.log.SetLevel(log.DebugLevel)
	s.log.SetOutput(os.Stderr)

	t.Run("assert each node's pod count is correct", func(t *testing.T) {
		for _, n := range nodes {
			assert.Len(t, n.Pods, podCountPerNode)
		}
	})
}
