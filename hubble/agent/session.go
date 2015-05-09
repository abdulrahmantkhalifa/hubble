package agent

import (
	"hubble"
)
const sessionQueueSize = 512
var sessions = make(map[string]chan *hubble.MessageCapsule)
