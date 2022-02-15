package amqpchannelmanager

import (
	"sync"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/streadway/amqp"
)

var (
	ChannelLock sync.Mutex
)

func PublishToJobsQueue(cntr controller.CntrInterface, job []byte) error {
	ChannelLock.Lock()
	for {
		if len(configuration.AmqpChannels) > 0 {
			currChannel := popOffAChannel()
			ChannelLock.Unlock()
			err := cntr.PublishToJobsQueue(currChannel, job)

			ChannelLock.Lock()
			configuration.AmqpChannels = append(configuration.AmqpChannels, currChannel)
			ChannelLock.Unlock()

			// TODO bump common and replace with commonUtil.MakeErr(err)
			return err
		}
	}
}

func popOffAChannel() amqp.Channel {
	currChannel := configuration.AmqpChannels[0]
	configuration.AmqpChannels = append(configuration.AmqpChannels[:0], configuration.AmqpChannels[1:]...)
	return currChannel
}
