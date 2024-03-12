package pconst

const (
	CustomAdminButtonGroupMsg        = 3001
	CustomAdminButtonConfigSignIn    = 3002
	CustomAdminButtonDistributeToken = 3003
	CustomAdminButtonExportMembers   = 3004
	CustomAdminWithdraw              = 3005

	CustomSubmitWallet = 1001
)

const (
	CmdStart         = "start"
	CmdHelp          = "help"
	CmdUserData      = "userdata"
	CmdGroupMsgStat  = "groupmsgstat"
	CmdSetScore      = "sendpoints"
	CmdConfigSignin  = "setcheckinrules"
	CmdDisableSignin = "disablecheckinrules"
	CmdWithdraw      = "withdraw"
	CmdSignIn        = "sign_in"
	CmdScore         = "score"
)

var GroupCmdList = []string{CmdSignIn, CmdScore}
var PrivateCmdList = []string{CmdStart, CmdHelp}

var CmdDesc = map[string]string{
	CmdScore:  "Check your score",
	CmdSignIn: "Sign in to get rewards",
	CmdStart:  "Start participating in an event",
	CmdHelp:   "Help",
}
