package service

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/tristan-club/cascade/model"
	"github.com/tristan-club/cascade/pconst"
	"github.com/tristan-club/cascade/pkg/bot/apiclient"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	"github.com/tristan-club/cascade/pkg/xrpc"
	"github.com/tristan-club/cascade/pkg/xrpc/xproto/taskmonitor_pb"
	he "github.com/tristan-club/kit/error"
	tlog "github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

func parseGroupUsername(input string) string {
	groupUsername := input
	if !strings.HasPrefix(groupUsername, "@") {
		groupUsername = "@" + groupUsername
	}
	return groupUsername
}

func parseUsername(input string) string {
	return strings.TrimPrefix(input, "@")
}

func getGroupIdByUsername(cli *apiclient.Client, groupUsername string) (int64, error) {
	id, err := rdb.Default().Get(context.Background(), pconst.RKPrefixGroupId+groupUsername).Int64()
	if err != nil && err != redis.Nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get redis data error", "error": err.Error()}).Send()
		return 0, err
	}

	if err == redis.Nil {
		chat, err := cli.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{SuperGroupUsername: groupUsername}})
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "get chat info config error", "error": err.Error()}).Send()
			return 0, fmt.Errorf("fetch group data error: %s", err.Error())
		}

		id = chat.ID

		err = rdb.Default().Set(context.Background(), pconst.RKPrefixGroupId+groupUsername, chat.ID, time.Minute).Err()
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "save group data cache error", "error": err.Error()}).Send()
		}
	}

	return id, nil
}

func checkIsGroupAdmin(ctx *Context, chatUsername string, chatId, userId int64) (bool, int64, error) {

	if chatId == 0 {
		var err error
		chatId, err = getGroupIdByUsername(ctx.Client, chatUsername)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "get chat id error", "error": err.Error()}).Send()
			return false, 0, err
		}
	}

	member, err := ctx.Client.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatId,
			UserID: userId,
		}})
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get chat member error", "error": err.Error()}).Send()
		return false, 0, fmt.Errorf("fetch user permission data error: %s", err.Error())
	}

	return member.IsAdministrator() || member.IsCreator(), chatId, nil
}

func HandleSetUserScore(ctx *Context) error {
	tlog.Info().Fields(map[string]interface{}{"action": "handle set user point", "ctx": ctx}).Send()

	if ctx.Update.CallbackData() != "" {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextDistributeTokenHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	param := strings.Fields(ctx.Update.Message.CommandArguments())
	if len(param) != 3 {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextDistributeTokenHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	groupUsername := parseGroupUsername(param[0])
	username := parseUsername(param[1])

	isAdmin, groupId, err := checkIsGroupAdmin(ctx, groupUsername, 0, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "check is group error", "error": err.Error()}).Send()
		return err
	} else if !isAdmin {
		tlog.Info().Fields(map[string]interface{}{"action": "permission refuse"}).Send()
		return he.NewStandardBusinessError(TextPermissionRefuse)
	}

	user, err := model.GetUserByUsername(username)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get user error", "error": err.Error()}).Send()
		return err
	} else if user == nil {
		tlog.Info().Fields(map[string]interface{}{"action": "get user not found"}).Send()
		return he.NewStandardBusinessError(TextUsernameNotFound)
	}

	amount, err := strconv.ParseInt(param[2], 10, 64)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "parse amount error", "error": err.Error()}).Send()
		return he.NewStandardBusinessError(fmt.Sprintf("invalid amount input: %s", param[2]))
	}

	score, err := rdb.Default().HGet(context.Background(), pconst.GetRKScore(groupId), strconv.FormatInt(user.OpenId, 10)).Int()
	if err != nil && err != redis.Nil {
		tlog.Error().Fields(map[string]interface{}{"action": "set score error", "error": err.Error()}).Send()
		return err
	}

	if amount < 0 && int(amount)+score < 0 {
		amount = -int64(score)
	}

	err = rdb.Default().HIncrBy(context.Background(), pconst.GetRKScore(groupId), strconv.FormatInt(user.OpenId, 10), amount).Err()
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "set score error", "error": err.Error()}).Send()
		return err
	}

	content := fmt.Sprintf("Operation Success!\n"+
		"Now %s score %d", username, score+int(amount))
	_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, content, nil, false)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	return nil
}

func HandleGroupMsgStat(ctx *Context) error {

	req := &taskmonitor_pb.GetUserMsgCountListReq{}
	tlog.Info().Fields(map[string]interface{}{"action": "handle group msg stat", "ctx": ctx}).Send()

	if ctx.Update.CallbackData() != "" {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextGroupMsgStatHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	param := strings.Fields(ctx.Update.Message.CommandArguments())
	if len(param) != 3 {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextGroupMsgStatHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	groupUsername := parseGroupUsername(param[0])

	var err error
	minMsgCount, err := strconv.ParseUint(param[2], 10, 64)
	if err != nil {
		tlog.Info().Fields(map[string]interface{}{"action": "invalid min msg count"}).Send()
		return he.NewStandardBusinessError(TextInvalidMinMsgCount)
	}
	if minMsgCount == 0 {
		minMsgCount = 1
	}

	req.MinMsgCount = uint32(minMsgCount)
	duration := strings.ToLower(param[1])
	if len(duration) < 2 {
		tlog.Info().Fields(map[string]interface{}{"action": "invalid duration input"}).Send()
		return he.NewStandardBusinessError(TextInvalidDuration)
	}

	lastC := []byte(duration)[len(duration)-1]
	digitBit, err := strconv.ParseInt(string([]byte(duration)[:len(duration)-1]), 10, 64)
	if err != nil {
		tlog.Info().Fields(map[string]interface{}{"action": "invalid digit input"}).Send()
		return he.NewStandardBusinessError(TextInvalidDuration)
	}

	var fromTimestamp, toTimestamp int64

	if lastC == 'd' || lastC == 'h' {
		toTimestamp = time.Now().Unix()
		if lastC == 'd' {
			fromTimestamp = time.Now().Add(-time.Hour * 24 * time.Duration(digitBit)).Unix()
		} else {
			fromTimestamp = time.Now().Add(-time.Hour * time.Duration(digitBit)).Unix()
		}
	} else {
		durationArray := strings.Split(duration, ":")
		if len(durationArray) != 2 {
			tlog.Info().Fields(map[string]interface{}{"action": "invalid duration input"}).Send()
			return he.NewStandardBusinessError(TextInvalidDuration)
		}

		fromTime, err := time.Parse(pconst.StatLayout, durationArray[0])
		if err != nil {
			tlog.Info().Fields(map[string]interface{}{"action": "invalid from time"}).Send()
			return he.NewStandardBusinessError(TextInvalidDuration)
		}
		toTime, err := time.Parse(pconst.StatLayout, durationArray[1])
		if err != nil {
			tlog.Info().Fields(map[string]interface{}{"action": "invalid to time"}).Send()
			return he.NewStandardBusinessError(TextInvalidDuration)
		}
		fromTimestamp = fromTime.Unix()
		toTimestamp = toTime.Unix()

	}

	if toTimestamp-fromTimestamp > int64(60*60*24*60) {
		tlog.Info().Fields(map[string]interface{}{"action": "duration out of range"}).Send()
		return he.NewStandardBusinessError("The maximum query period is 60 days.")
	}

	isAdmin, groupId, err := checkIsGroupAdmin(ctx, groupUsername, 0, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "check is group error", "error": err.Error()}).Send()
		return err
	} else if !isAdmin {
		tlog.Info().Fields(map[string]interface{}{"action": "permission refuse"}).Send()
		return he.NewStandardBusinessError(TextPermissionRefuse)
	}

	req = &taskmonitor_pb.GetUserMsgCountListReq{
		GroupId:       strconv.FormatInt(groupId, 10),
		OpenIds:       nil,
		Page:          1,
		Limit:         pconst.GroupMsgQueryLimit,
		MinMsgCount:   uint32(minMsgCount),
		FromTimestamp: fromTimestamp,
		ToTimestamp:   toTimestamp,
		WithUserData:  true,
	}

	mcs, err := xrpc.GetTaskMonitorClient().GetUserMsgCountList(context.Background(), req)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get user msg count error", "req": req, "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	content := "*Top 50:*\n\n"
	for k, mc := range mcs.List {
		if k == 50 {
			break
		}
		content += fmt.Sprintf("%d\\. [%s](tg://user?id=%s) : *%d*\n", k+1, mdparse.ParseV2(mc.Firstname), mc.OpenId, mc.Count)

	}

	tlog.Debug().Msgf("state content: %s", content)

	filePath := fmt.Sprintf("./%s-msg-%s.xlsx", groupUsername, time.Now().Format(pconst.StatLayout))

	info, err := os.Stat(filePath)
	if !(err == nil && info != nil && !info.IsDir()) {

		file := excelize.NewFile()
		defer func() {
			if err := file.Close(); err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "close file error", "error": err.Error()}).Send()
			}
			if err = os.RemoveAll(filePath); err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "remove file error", "error": err.Error()}).Send()
			}
		}()

		file.SetCellValue("Sheet1", "A1", "Username")
		file.SetCellValue("Sheet1", "B1", "Count")
		file.SetCellValue("Sheet1", "C1", "Firstname")
		file.SetCellValue("Sheet1", "D1", "Lastname")

		for i, v := range mcs.List {
			file.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), v.Username)
			file.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), v.Count)
			file.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), v.Firstname)
			file.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), v.Lastname)
		}

		err = file.SaveAs(filePath)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "save xlsx error", "error": err.Error()}).Send()
			return err
		}
	}

	_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, content, nil, true)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	_, err = ctx.Client.SendDocument(ctx.Update.SentFrom().ID, filePath)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send document error", "error": err.Error()}).Send()
		return err
	}

	tlog.Info().Fields(map[string]interface{}{"action": "get user msg count", "mcs": mcs}).Send()

	return nil
}

func HandleUserData(ctx *Context) error {
	tlog.Info().Fields(map[string]interface{}{"action": "handle user data stat", "ctx": ctx}).Send()

	if ctx.Update.CallbackData() != "" {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextUserDataHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	param := strings.Fields(ctx.Update.Message.CommandArguments())
	if len(param) != 2 {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextUserDataHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	groupUsername := parseGroupUsername(param[0])

	isAdmin, groupId, err := checkIsGroupAdmin(ctx, groupUsername, 0, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "check is group error", "error": err.Error()}).Send()
		return err
	} else if !isAdmin {
		tlog.Info().Fields(map[string]interface{}{"action": "permission refuse"}).Send()
		return he.NewStandardBusinessError(TextPermissionRefuse)
	}

	openIdList := make([]int64, 0)
	scoreMap := make(map[int64]int)
	var user *model.User

	if strings.ToLower(param[1]) == "all" {
		scores, err := rdb.Default().HGetAll(context.Background(), pconst.GetRKScore(groupId)).Result()
		if err != nil && err != redis.Nil {
			tlog.Error().Fields(map[string]interface{}{"action": "get score error", "error": err.Error()}).Send()
			return err
		}
		for k, v := range scores {
			openId, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "invalid score open id", "error": err.Error()}).Send()
			} else {
				openIdList = append(openIdList, openId)
				score, _ := strconv.ParseInt(v, 10, 64)
				scoreMap[openId] = int(score)
			}
		}
	} else {
		user, err = model.GetUserByUsername(parseUsername(param[1]))
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "get user error", "error": err.Error()}).Send()
			return err
		} else if user == nil {
			tlog.Info().Fields(map[string]interface{}{"action": "get user not found"}).Send()
			return he.NewStandardBusinessError(TextUsernameNotFound)
		}
		openIdList = append(openIdList, user.OpenId)
		score, err := rdb.Default().HGet(context.Background(), pconst.GetRKScore(groupId), strconv.FormatInt(user.OpenId, 10)).Int()
		if err != nil && err != redis.Nil {
			tlog.Error().Fields(map[string]interface{}{"action": "get score error", "error": err.Error()}).Send()
			return err
		}
		scoreMap[user.OpenId] = score
	}

	if user != nil {

		content := fmt.Sprintf(""+
			"*Username:* %s\n"+
			"*Firstname:* %s\n"+
			"*Lastname:* %s\n"+
			"*Score:* %d\n"+
			"*Wallet:* %s\n", user.Username, user.FirstName, user.LastName, scoreMap[user.OpenId], user.Addr)
		_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, mdparse.ParseV2(content), nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
			return he.NewStandardServerError(err)
		}
	} else {
		users, err := model.BatchGetUser(openIdList)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "get user error", "error": err.Error()}).Send()
			return err
		}

		content := fmt.Sprintf("Current sign\\-in user\\: %d", len(scoreMap))
		var filePath string
		if len(users) > 0 {
			filePath = fmt.Sprintf("./%s-user-%s.xlsx", groupUsername, time.Now().Format(pconst.StatLayout))
			file := excelize.NewFile()
			defer func() {
				if err := file.Close(); err != nil {
					tlog.Error().Fields(map[string]interface{}{"action": "close file error", "error": err.Error()}).Send()
				}
				if _, err = os.ReadDir(filePath); err != nil {
					tlog.Error().Fields(map[string]interface{}{"action": "remove file error", "error": err.Error()}).Send()
				}
			}()

			file.SetCellValue("Sheet1", "A1", "Username")
			file.SetCellValue("Sheet1", "B1", "Score")
			file.SetCellValue("Sheet1", "C1", "Firstname")
			file.SetCellValue("Sheet1", "D1", "Lastname")
			file.SetCellValue("Sheet1", "E1", "Addr")

			for i, v := range users {
				file.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), v.Username)
				file.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), scoreMap[v.OpenId])
				file.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), v.FirstName)
				file.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), v.LastName)
				file.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), v.Addr)
			}

			err = file.SaveAs(filePath)
			if err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "save xlsx error", "error": err.Error()}).Send()
				return err
			}
		}

		_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, content, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
			return he.NewStandardServerError(err)
		}

		if filePath != "" {
			_, err = ctx.Client.SendDocument(ctx.Update.SentFrom().ID, filePath)
			if err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "send document error", "error": err.Error()}).Send()
				return err
			}
		}
	}

	return nil

}

func HandleConfigSignin(ctx *Context) error {
	tlog.Info().Fields(map[string]interface{}{"action": "handle config signin", "ctx": ctx}).Send()

	if ctx.Update.CallbackData() != "" {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextConfigSigninRules, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	param := strings.Split(ctx.Update.Message.CommandArguments(), "|")

	if len(param) != 3 {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextConfigSigninRules, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	groupUsername := parseGroupUsername(param[0])

	amount, err := strconv.ParseInt(param[2], 10, 64)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "invalid signin reward input", "error": err.Error()}).Send()
		return err
	}

	isAdmin, groupId, err := checkIsGroupAdmin(ctx, groupUsername, 0, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "check is group error", "error": err.Error()}).Send()
		return err
	} else if !isAdmin {
		tlog.Info().Fields(map[string]interface{}{"action": "permission refuse"}).Send()
		return he.NewStandardBusinessError(TextPermissionRefuse)
	}

	sc, err := model.GetSignInConfig(groupId)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get signin config error", "error": err.Error()}).Send()
		return err
	}
	sc.CreatorId = ctx.GetOpenId()
	sc.Enable = true
	sc.Reward = int(amount)
	sc.Text = param[1]
	if err = model.SaveSignInConfig(sc); err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "sign in error", "error": err.Error()}).Send()
		return err
	}
	_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, TextOpSuccess, nil, false)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	return nil
}

func HandleDisableSignin(ctx *Context) error {

	groupUsername := parseGroupUsername(ctx.Update.Message.CommandArguments())

	isAdmin, groupId, err := checkIsGroupAdmin(ctx, groupUsername, 0, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "check is group error", "error": err.Error()}).Send()
		return err
	} else if !isAdmin {
		tlog.Info().Fields(map[string]interface{}{"action": "permission refuse"}).Send()
		return he.NewStandardBusinessError(TextPermissionRefuse)
	}

	sc, err := model.GetSignInConfig(groupId)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get signin config error", "error": err.Error()}).Send()
		return err
	}
	sc.Enable = false
	if err = model.SaveSignInConfig(sc); err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "sign in error", "error": err.Error()}).Send()
		return err
	}

	_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, TextOpSuccess, nil, false)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	return nil
}

func HandleWithdraw(ctx *Context) error {
	tlog.Info().Fields(map[string]interface{}{"action": "handle withdraw", "ctx": ctx}).Send()

	if ctx.Update.CallbackData() != "" {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextWithdrawHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	param := strings.Split(ctx.Update.Message.CommandArguments(), "|")

	if len(param) != 1 {
		_, err := ctx.Client.SendMsg(ctx.GetOpenId(), TextWithdrawHelp, nil, true)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send msg error", "error": err.Error()}).Send()
		}
		return err
	}

	groupUsername := parseGroupUsername(param[0])

	isAdmin, groupId, err := checkIsGroupAdmin(ctx, groupUsername, 0, ctx.GetOpenId())
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "check is group error", "error": err.Error()}).Send()
		return err
	} else if !isAdmin {
		tlog.Info().Fields(map[string]interface{}{"action": "permission refuse"}).Send()
		return he.NewStandardBusinessError(TextPermissionRefuse)
	}

	groupMemberCount, err := model.GetGroupMemberCount(groupId)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get group member error", "error": err.Error()}).Send()
		return err
	}

	claimedCount, err := model.GetGroupWithdraw(groupId)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get group claimed error", "error": err.Error()}).Send()
		return err
	}

	newlyAmount := uint64(groupMemberCount) - uint64(claimedCount)
	withdrawThreshold, _ := strconv.ParseInt(os.Getenv("WITHDRAW_THRESHOLD"), 10, 64)

	if newlyAmount == 0 || newlyAmount < uint64(withdrawThreshold) {
		tlog.Info().Fields(map[string]interface{}{"action": "claim not reach threshold"}).Send()
		content := fmt.Sprintf(""+
			"❌You do not have a sufficient number of new members to withdraw, the minimum withdrawal amount is: %d\n"+
			"Your current total number of members is: %d\n"+
			"You have already withdrawn an amount of: %d", withdrawThreshold, groupMemberCount, claimedCount)

		_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, content, nil, false)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
			return he.NewStandardServerError(err)
		}
		return nil
	}

	if err := model.AddGroupClaimed(groupId, newlyAmount); err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "add group claimed error", "error": err.Error()}).Send()
		return err
	}

	log.Info().Msgf("add group %d newly claimed request %d", groupId, newlyAmount)

	withdraw := model.NewWithdraw(groupId, ctx.GetOpenId(), ctx.Update.SentFrom().UserName, ctx.User.Addr, uint64(newlyAmount))
	err = model.AddWithdraw(nil, withdraw)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "save withdraw error", "error": err.Error()}).Send()
		return err
	}

	_, err = ctx.Client.SendMsg(ctx.Update.SentFrom().ID, TextSubmitSuccess, nil, false)
	if err != nil {
		tlog.Error().Fields(map[string]interface{}{"action": "send question error", "error": err.Error()}).Send()
		return he.NewStandardServerError(err)
	}

	return nil
}
