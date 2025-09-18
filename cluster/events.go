package cluster

import (
	"charry/event"
)

// 集群事件类型常量
const (
	// EventNodeAdded 节点添加事件
	EventNodeAdded = "cluster.node.added"

	// EventNodeUpdated 节点更新事件
	EventNodeUpdated = "cluster.node.updated"

	// EventNodeRemoved 节点删除事件
	EventNodeRemoved = "cluster.node.removed"

	// EventClusterChanged 集群整体变化事件
	EventClusterChanged = "cluster.changed"

	// EventClusterConnected 集群连接事件
	EventClusterConnected = "cluster.connected"

	// EventClusterDisconnected 集群断开连接事件
	EventClusterDisconnected = "cluster.disconnected"
)

// NodeEventData 节点事件数据
type NodeEventData struct {
	Node     *Node                  `json:"node"`     // 变更的节点信息
	OldNode  *Node                  `json:"old_node"` // 旧节点信息（仅在更新事件中有值）
	Action   string                 `json:"action"`   // 操作类型：add, update, remove
	Reason   string                 `json:"reason"`   // 变更原因
	Metadata map[string]interface{} `json:"metadata"` // 额外元数据
}

// ClusterEventData 集群事件数据
type ClusterEventData struct {
	Nodes        []*Node                `json:"nodes"`         // 当前所有节点
	AddedNodes   []*Node                `json:"added_nodes"`   // 新增节点
	UpdatedNodes []*Node                `json:"updated_nodes"` // 更新节点
	RemovedNodes []*Node                `json:"removed_nodes"` // 删除节点
	TotalCount   int                    `json:"total_count"`   // 节点总数
	Metadata     map[string]interface{} `json:"metadata"`      // 额外元数据
}

// CreateNodeAddedEvent 创建节点添加事件
func CreateNodeAddedEvent(node *Node, reason string) event.Event {
	data := &NodeEventData{
		Node:     node,
		Action:   "add",
		Reason:   reason,
		Metadata: make(map[string]interface{}),
	}

	return event.NewEvent(EventNodeAdded, "cluster-manager", data)
}

// CreateNodeUpdatedEvent 创建节点更新事件
func CreateNodeUpdatedEvent(newNode, oldNode *Node, reason string) event.Event {
	data := &NodeEventData{
		Node:     newNode,
		OldNode:  oldNode,
		Action:   "update",
		Reason:   reason,
		Metadata: make(map[string]interface{}),
	}

	return event.NewEvent(EventNodeUpdated, "cluster-manager", data)
}

// CreateNodeRemovedEvent 创建节点删除事件
func CreateNodeRemovedEvent(node *Node, reason string) event.Event {
	data := &NodeEventData{
		Node:     node,
		Action:   "remove",
		Reason:   reason,
		Metadata: make(map[string]interface{}),
	}

	return event.NewEvent(EventNodeRemoved, "cluster-manager", data)
}

// CreateClusterChangedEvent 创建集群变化事件
func CreateClusterChangedEvent(nodes, addedNodes, updatedNodes, removedNodes []*Node) event.Event {
	data := &ClusterEventData{
		Nodes:        nodes,
		AddedNodes:   addedNodes,
		UpdatedNodes: updatedNodes,
		RemovedNodes: removedNodes,
		TotalCount:   len(nodes),
		Metadata:     make(map[string]interface{}),
	}

	return event.NewEvent(EventClusterChanged, "cluster-manager", data)
}

// CreateClusterConnectedEvent 创建集群连接事件
func CreateClusterConnectedEvent(nodes []*Node) event.Event {
	data := &ClusterEventData{
		Nodes:      nodes,
		TotalCount: len(nodes),
		Metadata:   make(map[string]interface{}),
	}

	return event.NewEvent(EventClusterConnected, "cluster-manager", data)
}

// CreateClusterDisconnectedEvent 创建集群断开连接事件
func CreateClusterDisconnectedEvent(reason string) event.Event {
	data := map[string]interface{}{
		"reason":    reason,
		"timestamp": event.Event{}.Timestamp,
	}

	return event.NewEvent(EventClusterDisconnected, "cluster-manager", data)
}
