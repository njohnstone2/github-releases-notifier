package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

// SlackSender has the hook to send slack notifications.
type SlackSender struct {
	Client  *slack.Client
	Channel string
}

type slackPayload struct {
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
	Text      string `json:"text"`
}

// Send a notification with a formatted message build from the repository.
func (s *SlackSender) Send(repository Repository) error {
	headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(":trumpet: New release published! :trumpet:\n<%s|Release - %s>", repository.Release.URL.String(), repository.Release.Name), false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	attachment := slack.Attachment{
		Color: "#2eb886",
		Fields: []slack.AttachmentField{
			{
				Value: repository.Release.Description,
			},
		},
		FooterIcon: "https://static-00.iconduck.com/assets.00/github-icon-256x249-eb1fu3cu.png",
		Footer:     fmt.Sprintf("https://github.com/%s/%s", repository.Owner, repository.Name),
	}

	_, timestamp, err := s.Client.PostMessage(
		s.Channel,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionText("", false),
		slack.MsgOptionBlocks(
			headerSection,
		),
	)

	if err != nil {
		return err
	}
	fmt.Printf("Message sent at %s", timestamp)

	return nil
}
