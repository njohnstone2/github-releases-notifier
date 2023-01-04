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
	msg := []slack.MsgOption{}
	if repository.Release != nil {
		msg = s.buildReleaseMessage(repository)
	} else if repository.Tag != nil {
		msg = s.buildTagMessage(repository)
	} else {
		return fmt.Errorf("can't convert tag id to string: %v", repository.ID)
	}

	_, timestamp, err := s.Client.PostMessage(
		s.Channel,
		msg...,
	)

	if err != nil {
		return err
	}
	fmt.Printf("Message sent at %s", timestamp)

	return nil
}

func (s *SlackSender) buildReleaseMessage(r Repository) []slack.MsgOption {
	title := fmt.Sprintf("%s/%s:%s", r.Owner, r.Name, r.Release.Name)
	headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(":rocket: New release published! :rocket:\n<%s|Release - %s>", r.Release.URL.String(), title), false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	attachment := slack.Attachment{
		Color: "#2eb886",
		Fields: []slack.AttachmentField{
			{
				Value: r.Release.Description,
			},
		},
		FooterIcon: "https://static-00.iconduck.com/assets.00/github-icon-256x249-eb1fu3cu.png",
		Footer:     fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name),
	}

	return []slack.MsgOption{
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionText("", false),
		slack.MsgOptionBlocks(
			headerSection,
		),
	}
}

func (s *SlackSender) buildTagMessage(r Repository) []slack.MsgOption {
	url := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", r.Owner, r.Name, r.Tag.Name)
	attachment := slack.Attachment{
		Color: "#2eb886",
		Fields: []slack.AttachmentField{
			{
				Title: fmt.Sprintf("%s/%s", r.Owner, r.Name),
				Value: fmt.Sprintf(":label: <%s|%s> tag published! :label:", url, r.Tag.Name),
			},
		},
		FooterIcon: "https://static-00.iconduck.com/assets.00/github-icon-256x249-eb1fu3cu.png",
		Footer:     fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name),
	}

	return []slack.MsgOption{
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionText("", false),
	}
}
