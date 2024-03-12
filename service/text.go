package service

const (
	TextStart              = "🌟 Feel free to reach out to me anytime you need assistance with your groups. Simply type /help to discover all the amazing ways I can assist you.\nDon't miss out on the fun! \n\n🚀 Add me to your chat now! 🚀"
	TextHelp               = "Hey there! I'm Cascade, your ultimate group builder bot 🌟. I'm here to turbocharge your group's engagement and activity by helping you set up an awesome points system! Whether it's daily check-ins, monitoring message counts, or more, I've got you covered.\n\nFeel free to start exploring by hitting the buttons below:\n- /start: Initiates my services.\n- /help: Summons detailed information about what I can do.\n\nLet's make your group more lively together! 💥🤖"
	TextShouldSubmitWallet = "Oopsie-daisy! 🌼 It seems you haven't shared your wallet address with me yet. 💌 Please provide it first, or else I won't be able to send you those delightful rewards. "

	TextCheckInSuccess   = "Congratulations %s\\! You have successfully signed in for today\\. You\\'ve earned *%d %s* as a sign\\-in reward\\. Keep up the good work\\! 🎉🏆"
	TextAlreadyCheckIn   = "You have already signed in today\\. Please come back tomorrow for your next reward\\! 📅⏰"
	TextScore            = "Your score is %d"
	TextOpSuccess        = "Operation Success"
	TextInputAddress     = "Please input your ton wallet address"
	TextGroupMsgStatHelp = "" +
		"*Group Message Statistics*\n" +
		"To retrieve group message statistics\\, use the following command\\:\n\n" +
		"`\\/groupmsgstat \\{groupUsername\\} \\{period\\} \\{minimum message count\\}`\n\n" +
		"\\- groupUsername\\: The group\\'s username\n" +
		"\\- period\\: Specify the time period using \\xh\\' for hours or \\'xd\\' for days before the current time\n" +
		"\\- minimum message count\\: Users with more than this number of messages will be included in the statistics\n\n" +
		"*Example\\:*\nTo check messages in the @TristanClubofficial group from the past 24 hours with a minimum message count of 1\\:\n\n" +
		"`\\/groupmsgstat @TristanClubofficial 1d 1`\n\n" +
		"*Note\\:*\n" +
		"Make sure the bot is a member of the group specified by the group username you enter\\. Otherwise\\, it won\\'t be able to accurately collect statistics\\."

	TextUserDataHelp = "" +
		"*Member Info* \n\n" +
		"This command allows group administrators to export detailed member information\\, including their points and wallet addresses\\, into an Excel spreadsheet\\.\n\n" +
		"*Command*\n\n" +
		"`\\/userdata \\{groupUsername\\} \\{username\\}`\n" +
		"`\\/userdata \\{groupUsername\\} all`\n\n" +
		"*Example*:\n" +
		"`\\/userdata @yourGroup ireneee0419`\n" +
		"`\\/userdata @yourGroup all`"

	TextDistributeTokenHelp = "" +
		"*Send Points*\n\n" +
		"This feature allows group administrators to manually distribute points to users\\. " +
		"It\\'s especially useful for recording winners in group activities and keeping track of their earned points\\, " +
		"as well as monitoring the total points accumulated by all group members\\.\n\n" +
		"*Command:*\n\n" +
		"`\\/sendpoints \\{@groupUsername\\} \\{username\\} \\{amount\\}`\n\n" +
		"*Example:*\n\n" +
		"`\\/sendpoints @yourGroup ireneee0419 10`\n\n" +
		"*Text:*\n\n" +
		"To send points\\, please use the following command format\\:\n\n" +
		"`\\/sendpoints \\{@groupUsername\\} \\{username\\} \\{amount\\}`\n\n" +
		"*For example\\:*\n\n" +
		"`\\/sendpoints @yourGroup ireneee0419 10`\n" +
		"`\\/sendpoints @yourGroup ireneee0419 \\-10`\n\n" +
		"Feel free to use this feature to record winners\\' points and check the cumulative points earned by all group members\\.  💰📊🏆"

	TextConfigSigninRules = "" +
		"*Configure Daily Check\\-Ins*\n\n" +
		"Boost member engagement in your group by setting up daily check\\-in rules\\. " +
		"Reward members with points for each check\\-in and let the bot send a customized message while mentioning the member\\.\n\n" +
		"*Commands\\:*\n\n" +
		"*1\\.Set Check\\-In Rules:*\n" +
		"`\\/setcheckinrules \\{groupUsername\\}\\|\\{text\\}\\|\\{points\\}`\n\n" +
		"\\- groupUsername\\: Your group's username\\.\n" +
		"\\- text\\: Message to send when members check in\\.\n" +
		"\\- points\\: Points awarded for each daily check\\-in\\.\n\n" +
		"*Example\\:*\n" +
		"To reward @yourGroup members with 10 points and send \"Thank you for checking in today\\!\" when they use the check\\-in command\\:\n" +
		"`\\/setcheckinrules @yourGroup|Thank you for checking in today\\!\\|10`\n\n" +
		"*2\\.Disable Check\\-In Rules\\:*\n" +
		"`\\/disablecheckinrules \\{groupUsername\\}`\n\n" +
		"\\- groupUsername\\: Your group's username\\.\n\n" +
		"Example\\:\n" +
		"To disable check\\-in rules for @yourGroup\\:\n" +
		"`\\/disablecheckinrules @yourGroup`\n\n" +
		"*Note\\:*\n" +
		"\\- Make sure the bot is added to your group before configuring check\\-in rules\\.\n" +
		"\\- Each member can only check in once per day\\.\n" +
		"\\- Setting and disabling check\\-in rules are exclusive to group administrators\\."

	TextWithdrawHelp = "" +
		"*Withdraw Tokens*\n\n" +
		"This feature allows group administrators to withdraw CSC tokens from centralized account to your TON blockchain wallet\\. " +
		"CSC tokens can be used to unlock premium features we may introduce in the future\\. \n\n" +
		"*Command:*\n" +
		"`\\/withdraw \\{groupUsername\\}`\n\n" +
		"*Example\\:*\n" +
		"`\\/withdraw @yourGroup`\n\n" +
		"You can use this feature to access your CSC token balance and\\, if needed\\, withdraw tokens to your TON blockchain wallet\\. 💰🔗🚀"

	TextSubmitSuccess = "Your request has been submitted. We will process your on-chain withdrawal within 48 hours."

	TextWithdrawTxSuccess = "🎉*Congratulations\\!*\n" +
		"Your withdrawal transaction has been successfully processed\\.\n" +
		"Click the link [*Transaction Detail*](%s) to view details\\."
)

const (
	TextInvalidCmdParam      = "invalid command param"
	TextInvalidDuration      = "invalid query period input"
	TextInvalidMinMsgCount   = "invalid minimum number of messages"
	TextUsernameNotFound     = "The data corresponding to the given username was not found.\nPerhaps the user you entered has never checked in, or they are no longer in the group."
	TextPermissionRefuse     = "You must first add the bot to your group. Please type /start to begin."
	TextCmdOnlyEffectInGroup = "This command can only take effect in group chat"
)

const (
	IButtonStart = "🚀I'm Ready to Begin!🚀"

	RButtonMyAccount = "📕My Account"
)
