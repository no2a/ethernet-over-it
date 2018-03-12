package main

import (
	"encoding/base64"
	"log"

	"github.com/nlopes/slack"
)

func doSlackClient(token string, channel string, toTAP chan []byte, fromTAP chan []byte) {
	api := slack.New(token)

	cs, err := api.GetChannels(true)
	if err != nil {
		log.Println(err)
	}
	channelID := ""
	for _, c := range cs {
		if c.Name == channel {
			channelID = c.ID
		}
	}
	if channelID == "" {
		log.Printf("channel %s did not found", channel)
		return
	}
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				packet, _ := base64.StdEncoding.DecodeString(ev.Text)
				toTAP <- packet
			}
		case packet := <-fromTAP:
			text := base64.StdEncoding.EncodeToString(packet)
			rtm.SendMessage(rtm.NewOutgoingMessage(text, channelID))
		}
	}
}
