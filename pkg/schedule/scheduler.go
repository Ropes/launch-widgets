package schedule

import (
	"sort"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Scheduler manages the Pods which need to be scheduled
type Scheduler struct {
	podsToStart map[uuid.UUID]*Pod

	// PodPriorityRecovery is a negative value which decrements the
	// invoiced value of the Pod, to a higher effective priority.
	podPriorityRecovery int

	log *log.Logger
}

func NewScheduler(pods []*Pod, priorityRecovery int) *Scheduler {
	podMap := make(map[uuid.UUID]*Pod)
	for _, p := range pods {
		podMap[p.ID] = p
	}
	return &Scheduler{
		podsToStart:         podMap,
		podPriorityRecovery: priorityRecovery,
		log:                 log.StandardLogger(),
	}
}

// sortPods of Scheduler's list into prioritized list
func (s *Scheduler) sortPods() []*Pod {
	podSlice := make([]*Pod, 0)

	for _, v := range s.podsToStart {
		podSlice = append(podSlice, v)
	}
	sort.Sort(ByPriority(podSlice))
	return podSlice
}

// AddPodsToSchedule to the priority queue for scheduling.
func (s *Scheduler) AddPodsToSchedule(pods []*Pod) {
	for _, p := range pods {
		s.podsToStart[p.ID] = p
	}
}

// AdjustPodPriority decrements priority by the PodPriorityRecovery
// value, but not lower than 0.
func (s *Scheduler) AdjustPodPriority() {
	for _, p := range s.podsToStart {
		p.Priority += s.podPriorityRecovery
		if p.Priority < 0 {
			p.Priority = 0
		}
	}
}

// ScheduleNodes accepts the slice of Nodes, and will attempt schedule Pods
// to them. First sorting the Pods by their priority, then iterating until
// all pods scheduled, or Nodes have no more capacity.
func (s *Scheduler) ScheduleNodes(nodes []*Node) {
	s.log.WithFields(log.Fields{"podCount": len(s.podsToStart), "nodeCount": len(nodes)}).Info("scheduling pods to nodes")
	pods := s.sortPods()
	nodeI := 0 // Starting node index for checking pod acceptance criteria.
	for _, p := range pods {
		s.log.WithFields(log.Fields{
			"pod": p.String(),
		}).Info("pod to be scheduled")
		for i := 0; i < len(nodes); i++ { //Iterate over all nodes
			// Begin iteration at nodeI to ensure even distribution
			n := nodes[(i+nodeI)%len(nodes)]
			s.log.WithFields(log.Fields{"node": n.Labels["name"]}).Debug("searching for capacity on node")
			nodeCapacity := n.Capacity()

			// Check Node's capacity, if it is higher than the Pod's
			// requirements, start Pod on Node.
			if nodeCapacity.greaterThan(&p.CR) {

				// TODO: handle additional constraints
				err := n.Start(p)
				if err != nil {
					s.log.Errorf("error scheduling pod to node[%v]: %v", p, n)
					continue
				}
				delete(s.podsToStart, p.ID) // Remove Pod from Scheduler's list
				s.log.WithFields(log.Fields{"pod": p.String(), "node": n.Labels["name"]}).Info("pod scheduled successfully")
				break // Move to next Pod for assignment
			}
		}
		nodeI++ // Increment node starting index
	}
	s.log.WithField("pods unscheduled", len(s.podsToStart)).Info("scheduling completed")
}
