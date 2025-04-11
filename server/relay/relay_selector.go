package relay

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// RelayNode 中继节点
type RelayNode struct {
	NodeID       string
	ExternalIP   string
	ExternalPort int
	Load         int
	Latency      time.Duration
	Bandwidth    int
	LastChecked  time.Time
}

// RelaySelector 中继节点选择器
type RelaySelector struct {
	nodes map[string]*RelayNode
	mu    sync.RWMutex
}

// NewRelaySelector 创建中继节点选择器
func NewRelaySelector() *RelaySelector {
	return &RelaySelector{
		nodes: make(map[string]*RelayNode),
	}
}

// AddNode 添加中继节点
func (s *RelaySelector) AddNode(node *RelayNode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[node.NodeID] = node
}

// RemoveNode 移除中继节点
func (s *RelaySelector) RemoveNode(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nodes, nodeID)
}

// UpdateNodeStatus 更新节点状态
func (s *RelaySelector) UpdateNodeStatus(nodeID string, load int, latency time.Duration, bandwidth int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.nodes[nodeID]
	if !exists {
		return errors.New("节点不存在")
	}

	node.Load = load
	node.Latency = latency
	node.Bandwidth = bandwidth
	node.LastChecked = time.Now()

	return nil
}

// SelectBestNode 选择最佳中继节点
func (s *RelaySelector) SelectBestNode(sourceNodeID, targetNodeID string) (*RelayNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 过滤掉源节点和目标节点
	var candidates []*RelayNode
	for id, node := range s.nodes {
		if id != sourceNodeID && id != targetNodeID {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return nil, errors.New("没有可用的中继节点")
	}

	// 根据负载、延迟和带宽计算得分
	type scoredNode struct {
		node  *RelayNode
		score float64
	}

	var scoredNodes []scoredNode
	for _, node := range candidates {
		// 计算得分，负载越低、延迟越低、带宽越高，得分越高
		loadScore := 100.0 / float64(node.Load+1)
		latencyScore := 100.0 / float64(node.Latency.Milliseconds()+1)
		bandwidthScore := float64(node.Bandwidth) / 100.0

		// 综合得分，可以根据实际情况调整权重
		score := loadScore*0.4 + latencyScore*0.3 + bandwidthScore*0.3

		scoredNodes = append(scoredNodes, scoredNode{node, score})
	}

	// 按得分排序
	sort.Slice(scoredNodes, func(i, j int) bool {
		return scoredNodes[i].score > scoredNodes[j].score
	})

	// 返回得分最高的节点
	return scoredNodes[0].node, nil
}

// SelectRandomNode 随机选择中继节点
func (s *RelaySelector) SelectRandomNode(sourceNodeID, targetNodeID string) (*RelayNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 过滤掉源节点和目标节点
	var candidates []*RelayNode
	for id, node := range s.nodes {
		if id != sourceNodeID && id != targetNodeID {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return nil, errors.New("没有可用的中继节点")
	}

	// 随机选择一个节点
	return candidates[rand.Intn(len(candidates))], nil
}
