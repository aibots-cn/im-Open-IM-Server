/*
** description("").
** copyright('OpenIM,www.OpenIM.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/5/13 10:33).
 */
package push

import (
	"OpenIM/pkg/common/config"
	"OpenIM/pkg/common/constant"
	kfk "OpenIM/pkg/common/kafka"
	"OpenIM/pkg/common/log"
	"OpenIM/pkg/common/tracelog"
	pbChat "OpenIM/pkg/proto/msg"
	pbPush "OpenIM/pkg/proto/push"
	"OpenIM/pkg/utils"
	"context"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
)

type ConsumerHandler struct {
	pushConsumerGroup *kfk.MConsumerGroup
	pusher            *Pusher
}

func NewConsumerHandler(pusher *Pusher) *ConsumerHandler {
	var consumerHandler ConsumerHandler
	consumerHandler.pusher = pusher
	consumerHandler.pushConsumerGroup = kfk.NewMConsumerGroup(&kfk.MConsumerGroupConfig{KafkaVersion: sarama.V2_0_0_0,
		OffsetsInitial: sarama.OffsetNewest, IsReturnErr: false}, []string{config.Config.Kafka.Ms2pschat.Topic}, config.Config.Kafka.Ms2pschat.Addr,
		config.Config.Kafka.ConsumerGroupID.MsgToPush)
	return &consumerHandler
}

func (c *ConsumerHandler) handleMs2PsChat(ctx context.Context, msg []byte) {
	log.NewDebug("", "msg come from kafka  And push!!!", "msg", string(msg))
	msgFromMQ := pbChat.PushMsgDataToMQ{}
	if err := proto.Unmarshal(msg, &msgFromMQ); err != nil {
		log.Error("", "push Unmarshal msg err", "msg", string(msg), "err", err.Error())
		return
	}
	pbData := &pbPush.PushMsgReq{
		MsgData:  msgFromMQ.MsgData,
		SourceID: msgFromMQ.SourceID,
	}
	sec := msgFromMQ.MsgData.SendTime / 1000
	nowSec := utils.GetCurrentTimestampBySecond()
	if nowSec-sec > 10 {
		return
	}
	tracelog.SetOperationID(ctx, "")
	var err error
	switch msgFromMQ.MsgData.SessionType {
	case constant.SuperGroupChatType:
		err = c.pusher.MsgToSuperGroupUser(ctx, pbData.SourceID, pbData.MsgData)
	default:
		err = c.pusher.MsgToUser(ctx, pbData.SourceID, pbData.MsgData)
	}
	if err != nil {
		log.NewError("", "push failed", *pbData)
	}
}
func (ConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (c *ConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.NewDebug("", "kafka get info to mysql", "msgTopic", msg.Topic, "msgPartition", msg.Partition, "msg", string(msg.Value))
		ctx := c.pushConsumerGroup.GetContextFromMsg(msg)
		c.handleMs2PsChat(ctx, msg.Value)
		sess.MarkMessage(msg, "")
	}
	return nil
}