package main

import (
	"time"

	"github.com/ropes/launch-widgets/pkg/schedule"
)

func main() {
	podsToSchedule := schedule.NPods(30)
	nodes := schedule.NNodes(3)
	s := schedule.NewScheduler(podsToSchedule, -300)

	for {
		s.AdjustPodPriority()
		s.ScheduleNodes(nodes)
		podsToSchedule = []*schedule.Pod{}

		// Apply priority cost to running Pods on each Node
		schedule.NodesApplyCost(nodes)
		// Expire Pods with priority level above 50 on each Node
		podsToReshedule := schedule.ExpireLowPriorityPods(nodes, 500)
		s.AddPodsToSchedule(podsToReshedule)

		time.Sleep(5 * time.Second)
	}

}
