package performs

import (
	"context"
	"fmt"
	"rela_recommend/factory"
	"rela_recommend/log"
	"time"

	// "context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2write "github.com/influxdata/influxdb-client-go/v2/api/write"
)

type writeItem struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

var writeItemChan chan *writeItem = make(chan *writeItem, 10000)

// 批量写入买点，返回未写成功数据 和 错误信息
func writeBatchPoints(org string, bucket string, points []*influxdb2write.Point) ([]*influxdb2write.Point, error) {
	var noWritePoints = make([]*influxdb2write.Point, 0)
	var writeErr error
	if len(points) > 0 {
		if factory.InfluxdbClient != nil && len(factory.InfluxdbClient.ServerURL()) > 0 {
			writer := factory.InfluxdbClient.WriteAPIBlocking(org, bucket)
			if writeErr := writer.WritePoint(context.Background(), points...); writeErr != nil {
				log.Warnf("influxdb write err %s", writeErr.Error())
				noWritePoints = points
			} else {
				log.Infof("influxdb write len %d", len(points))
			}
		}
	}
	return noWritePoints, writeErr
}

func BeginWatching(org string, bucket string) {
	go func() {
		var points = []*influxdb2write.Point{}
		for { // 消费打点（10条写入一次 或 100毫秒写一次）
			select {
			case item := <-writeItemChan:
				points = append(points, influxdb2.NewPoint(item.Measurement, item.Tags, item.Fields, item.Time))
				if len(points) >= 20 {
					points, _ = writeBatchPoints(org, bucket, points)
				}
			case <-time.After(time.Millisecond * 300):
				points, _ = writeBatchPoints(org, bucket, points)
			}
		}
		// close(writeItemChan)
	}()
	fmt.Println("write performs to db begin...")
}
