package cluster

type Node struct {
	Id      string
	Name    string
	Type    int         // 节点类型
	LanAddr string      // 内网地址
	WanAddr string      // 公网地址
	Weights int         // 权重
	Data    interface{} // 节点数据
}
