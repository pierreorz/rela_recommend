package behavior

func (self *UserBehavior) GetUserInteracts() *Behavior {
	return self.Gets("userinfo:exposure_new", "userinfo:wink")
}

func (self *UserBehavior) GetUserPageViews() *Behavior {
	return self.Gets("userinfo:exposure_new")
}

func (self *UserBehavior) GetUserWinks() *Behavior {
	return self.Gets("userinfo:wink")
}

func (self *UserBehavior) GetSendMessages() *Behavior {
	return self.Gets("message.im:send_msg")
}
