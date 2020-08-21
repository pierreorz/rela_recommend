package performs

import (
	"context"
	"fmt"
	"rela_recommend/factory"
	"rela_recommend/log"
	"time"

	// "context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type writeItem struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

var writeItemChan chan *writeItem = make(chan *writeItem, 10000)

func BeginWatching(org string, bucket string) {
	go func() {
		for {
			select {
			case item := <-writeItemChan:
				point := influxdb2.NewPoint(item.Measurement, item.Tags, item.Fields, item.Time)
				if factory.InfluxdbClient != nil && len(factory.InfluxdbClient.ServerURL()) > 0 {
					writer := factory.InfluxdbClient.WriteAPIBlocking(org, bucket)
					if writeErr := writer.WritePoint(context.Background(), point); writeErr != nil {
						log.Warn("influxdb write err %s", writeErr.Error())
					}
				}
			case <-time.After(time.Second):
				log.Warn("influxdb write no data")
			}
		}
		// close(writeItemChan)
	}()
	fmt.Println("write performs to db begin...")
}
