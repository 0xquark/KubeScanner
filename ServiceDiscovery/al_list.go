package main

type ApplicationDiscoveryListItem struct {
	Discovery  ApplicationLayerDiscovery
	Reqirement string
}

var ApplicationDiscoveryList = []ApplicationDiscoveryListItem{
	{
		Discovery:  &KubeletDiscovery{},
		Reqirement: string(HTTP),
	},
	//	{
	//		Discovery:  &MysqlDiscovery{},
	//		Reqirement: string(TCP),
	//	},
}
