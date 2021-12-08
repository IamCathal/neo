package graphing

import (
	"github.com/go-echarts/go-echarts/charts"
	"github.com/iamcathal/neo/services/crawler/controller"
)

type graphConfig struct {
	links []charts.GraphLink
	nodes []charts.GraphNode
}

func GenerateGraph(cntr controller.CntrInterface, crawlID string) error {
	return GetAllUsersGraphableData(cntr, crawlID)
}

func GetAllUsersGraphableData(cntr controller.CntrInterface, crawlID string) error {
	// crawlingStatus, err := cntr.GetCrawlingStatsFromDataStore(crawlID)
	// if err != nil {
	// 	return err
	// }
	// totalUsersToCrawl := crawlingStatus.TotalUsersToCrawl
	// usersCrawled := crawlingStatus.UsersCrawled

	// var waitGroup sync.WaitGroup
	// var jobMutex sync.Mutex
	// var usersCrawledMutex sync.Mutex

	// workerConfig := graphWorkerConfig{
	// 	wg:                &waitGroup,
	// 	jobMutex:          &jobMutex,
	// 	usersCrawledMutex: &usersCrawledMutex,
	// 	TotalUsersToCrawl: 0,
	// 	UsersCrawled:      0,
	// }

	return nil
}

func GenerateGraphData(gConfig graphConfig) {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(
		charts.TitleOpts{Title: "graph title there"},
		charts.InitOpts{Width: "1800px", Height: "1080px"})

	graph.Add("graph", gConfig.nodes, gConfig.links,
		charts.GraphOpts{Layout: "force", Roam: true, Force: charts.GraphForce{Repulsion: 38, Gravity: 0.14}, FocusNodeAdjacency: true},
		charts.EmphasisOpts{Label: charts.LabelTextOpts{Show: true, Position: "left", Color: "black"}},
		charts.LineStyleOpts{Width: 1, Color: "#b5b5b5"},
	)

	// Render to temp file and then take the JSON from this
}

func SaveGraphDataToDataStore() {

}
