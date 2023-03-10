package sub

import (
	"github.com/cosmos/cosmos-sdk/pubsub"
	oTypes "github.com/cosmos/cosmos-sdk/x/oracle/types"

	"github.com/bnb-chain/node/plugins/bridge"
)

func SubscribeCrossTransferEvent(sub *pubsub.Subscriber) error {
	err := sub.Subscribe(bridge.CrossTransferTopic, func(event pubsub.Event) {
		switch event := event.(type) {
		case pubsub.CrossTransferEvent:
			crossTransferEvent := event
			if stagingArea.CrossTransferData == nil {
				stagingArea.CrossTransferData = make([]pubsub.CrossTransferEvent, 0, 1)
			}
			stagingArea.CrossTransferData = append(stagingArea.CrossTransferData, crossTransferEvent)

		default:
			sub.Logger.Info("unknown event type")
		}
	})
	return err
}

func SubscribeOracleEvent(sub *pubsub.Subscriber) error {

	err := sub.Subscribe(oTypes.Topic, func(event pubsub.Event) {
		switch event := event.(type) {
		case oTypes.CrossAppFailEvent:
			crossFailEvent := event
			sub.Logger.Info("do have crossFailEvent")

			// no need to publish into CrossTransferData if no balance change.
			if crossFailEvent.RelayerFee > 0 {
				if stagingArea.CrossTransferData == nil {
					stagingArea.CrossTransferData = make([]pubsub.CrossTransferEvent, 0, 1)
				}
				stagingArea.CrossTransferData = append(stagingArea.CrossTransferData, pubsub.CrossTransferEvent{
					TxHash:     crossFailEvent.TxHash,
					ChainId:    crossFailEvent.ChainId,
					Type:       bridge.CrossAppFailedType,
					RelayerFee: crossFailEvent.RelayerFee,
					From:       crossFailEvent.From,
				})
			}
		default:
			sub.Logger.Info("unknown event type")
		}
	})
	return err
}

func commitCrossTransfer() {
	if len(stagingArea.CrossTransferData) > 0 {
		toPublish.EventData.CrossTransferData = append(toPublish.EventData.CrossTransferData, stagingArea.CrossTransferData...)
	}
}
