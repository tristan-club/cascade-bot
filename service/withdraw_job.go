package service

import (
	"fmt"
	"github.com/tristan-club/cascade/model"
	"github.com/tristan-club/cascade/pkg/util"
	"github.com/tristan-club/kit/bignum"
	"github.com/tristan-club/kit/config"
	"github.com/tristan-club/kit/dingding"
	tlog "github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"math/big"
	"os"
	"runtime/debug"
	"time"
)

func DoWithdrawTxJob() error {

	defer func() {
		if err := recover(); err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "panic error", "error": err}).Send()
			dingding.Default().SendTextMessage(fmt.Sprintf("fission-bot user notify job panic error: %s", util.FastMarshal(err)), nil, false)
			debug.PrintStack()
		}
	}()

	err := doWithdrawTxJob()
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "do with draw job error", "error": err.Error()}).Send()
		dingding.Default().SendTextMessage(fmt.Sprintf("withdraw job get error: %s", err.Error()), nil, !config.EnvIsDev())
	}

	return nil
}

func doWithdrawTxJob() error {
	withdraws, err := model.GetShouldProcessWithdrawList()
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get should process withdraw error", "error": err.Error()}).Send()
		return err
	} else if len(withdraws) == 0 {
		return nil
	}

	withdrawStartId := withdraws[0].Id
	withdrawEndId := withdraws[len(withdraws)-1].Id

	tlog.Info().Fields(map[string]interface{}{"action": "start handle withdraw job", "withdraws": withdraws}).Send()

	userAddressMap := make(map[int64]string, 0)
	addressAmountMap := make(map[string]uint64, 0)

	ids := make([]uint, 0)

	for _, v := range withdraws {
		userAddressMap[v.OpenId] = v.Address
		addressAmountMap[v.Address] += v.Amount
		ids = append(ids, v.Id)
	}

	var addressList []string
	var amountList []string

	for address, amount := range addressAmountMap {
		amountBig := new(big.Int).SetUint64(amount)
		bignum.AddDecimal(amountBig, 9, 10)

		addressList = append(addressList, address)
		amountList = append(amountList, amountBig.String())
	}

	var txHash string
	dingText := fmt.Sprintf(`
## Cascade Daily Withdraw Request
Id range: %d-%d

### List
*address: amount*

`, withdrawStartId, withdrawEndId)
	claimList := ""
	for address, amount := range addressAmountMap {
		claimList += fmt.Sprintf("- %s: %d\n", address, amount)
	}
	dingText += claimList
	dingText += fmt.Sprintf(`

### AutoWithdrawParam
**addressList**:

%s

**amountList**:

%s`, util.FastMarshal(addressList), util.FastMarshal(amountList))

	if os.Getenv("AUTO_WITHDRAW") != "1" {
		err = dingding.Default().SendMarkdownMessage("Fission bot withdraw job", dingText, nil, false)
		if err != nil {
			return err
		}
	} else {
		tlog.Info().Msgf("start process withdraw tx, addressList: %v, amountList: %v", addressList, amountList)

		//		//txResp := &controller_pb.Tx{TxHash: "MockTxHash"}
		//		cli := tokentransfer.NewClient(a.ChainId, GetTokenTransferMgr(a.ChainId))
		//		txResp, err := cli.BatchTransfer(appMgr.App.OperatorUserId,
		//			appMgr.App.OperatorAddress,
		//			appMgr.App.OperatorPin,
		//			a.ContrAddr,
		//			addressList, amountList)
		//		if err != nil {
		//			tlog.Error().Fields(map[string]interface{}{"action": "transfer error", "error": err.Error()}).Send()
		//			return err
		//		}
		//
		//		txHash = txResp.TxHash
		//
		//		tlog.Info().Msgf("withdraw job tx success, txHash: %s, addressList: %s", txResp.TxHash, addressList)
		//
		//		dingText += fmt.Sprintf(`
		//
		//`, chain_info.GetExplorerTargetUrl(a.ChainId, txHash, chain_info.ExplorerTargetTransaction))
		//
		//		err = dingding.Default().SendMarkdownMessage("Fission bot withdraw job", dingText, nil, false)
		//		if err != nil {
		//			return err
		//		}
	}

	if err = model.BatchUpdateWithdraw(ids, map[string]interface{}{
		"is_proceed": 1,
		"tx_hash":    txHash,
	}); err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "update withdraw data error", "error": err.Error()}).Send()
		return fmt.Errorf("save withdraw info error: %s", err.Error())
	}

	go func(userAddressMap map[int64]string, txHash string) {
		if txHash == "" {
			return
		}

		for userId, _ := range userAddressMap {
			_, err = defaultManager().Client.SendMsg(userId,
				fmt.Sprintf(TextWithdrawTxSuccess,
					mdparse.ParseV2(fmt.Sprintf("https://tonviewer.com/transaction/%s", txHash))),
				nil, true)
			if err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error(), "userId": userId}).Send()
			}
			time.Sleep(time.Millisecond * 300)
		}

	}(userAddressMap, txHash)

	return nil
}
