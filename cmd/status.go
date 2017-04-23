// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var statusOpts struct {
	statusID int
	unset    bool

	// The following fields are used for the post/toot command
	visibility  string
	sensitive   bool
	spoiler     string
	inReplyToID int
	filePath    string
	mediaIDs    string

	// Used for several subcommands to limit the number of results
	limit uint
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Subcommands
	statusCmd.AddCommand(statusSubcommands...)

	// Global flags
	statusCmd.PersistentFlags().IntVarP(&statusOpts.statusID, "status-id", "s", 0, "Status ID number")
	statusCmd.PersistentFlags().UintVarP(&statusOpts.limit, "limit", "l", 0, "Limit number of results")

	statusCmd.MarkPersistentFlagRequired("status-id")

	// Subcommand flags
	statusReblogSubcommand.Flags().BoolVar(&statusOpts.unset, "unset", false, "Unreblog the status")
	statusFavouriteSubcommand.Flags().BoolVar(&statusOpts.unset, "unset", false, "Remove the status from the favourites")
	statusPostSubcommand.Flags().BoolVar(&statusOpts.sensitive, "sensitive", false, "Mark post as sensitive (NSFW)")
	statusPostSubcommand.Flags().StringVar(&statusOpts.visibility, "visibility", "", "Visibility (direct|private|unlisted|public)")
	statusPostSubcommand.Flags().StringVar(&statusOpts.spoiler, "spoiler", "", "Spoiler warning (CW)")
	statusPostSubcommand.Flags().StringVar(&statusOpts.mediaIDs, "media-ids", "", "Comma-separated list of media IDs")
	statusPostSubcommand.Flags().StringVarP(&statusOpts.filePath, "file", "f", "", "File name")
	statusPostSubcommand.Flags().IntVarP(&statusOpts.inReplyToID, "in-reply-to", "r", 0, "Status ID to reply to")
}

// statusCmd represents the status command
// This command does nothing without a subcommand
var statusCmd = &cobra.Command{
	Use:     "status --status-id ID subcommand",
	Aliases: []string{"st"},
	Short:   "Get status details",
	//Long:    `TBW...`, // TODO
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This is common to status and all status subcommands but "post"
		if statusOpts.statusID < 1 && cmd.Name() != "post" {
			return errors.New("missing status ID")
		}
		if err := madonInit(true); err != nil {
			return err
		}
		return nil
	},
}

var statusSubcommands = []*cobra.Command{
	&cobra.Command{
		Use:     "show",
		Aliases: []string{"display"},
		Short:   "Get the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "context",
		Short: "Get the status context",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "card",
		Short: "Get the status card",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "reblogged-by",
		Short: "Display accounts which reblogged the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "favourited-by",
		Aliases: []string{"favorited-by"},
		Short:   "Display accounts which favourited the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "Delete the status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusSubcommandRunE(cmd.Name(), args)
		},
	},
	statusReblogSubcommand,
	statusFavouriteSubcommand,
	statusPostSubcommand,
}

var statusReblogSubcommand = &cobra.Command{
	Use:     "boost",
	Aliases: []string{"reblog"},
	Short:   "Boost (reblog) or unreblog the status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusFavouriteSubcommand = &cobra.Command{
	Use:     "favourite",
	Aliases: []string{"favorite", "fave"},
	Short:   "Mark/unmark the status as favourite",
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

var statusPostSubcommand = &cobra.Command{
	Use:     "post",
	Aliases: []string{"toot", "pouet"},
	Short:   "Post a message (same as 'madonctl toot')",
	Example: `  madonctl status post --spoiler Warning "Hello, World"
  madonctl status toot --sensitive --file image.jpg Image
  madonctl status post --media-ids ID1,ID2,ID3 Image`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return statusSubcommandRunE(cmd.Name(), args)
	},
}

func statusSubcommandRunE(subcmd string, args []string) error {
	opt := statusOpts

	var obj interface{}
	var err error

	switch subcmd {
	case "show":
		var status *madon.Status
		status, err = gClient.GetStatus(opt.statusID)
		obj = status
	case "context":
		var context *madon.Context
		context, err = gClient.GetStatusContext(opt.statusID)
		obj = context
	case "card":
		var context *madon.Card
		context, err = gClient.GetStatusCard(opt.statusID)
		obj = context
	case "reblogged-by":
		var accountList []madon.Account
		accountList, err = gClient.GetStatusRebloggedBy(opt.statusID)
		if opt.limit > 0 {
			accountList = accountList[:opt.limit]
		}
		obj = accountList
	case "favourited-by":
		var accountList []madon.Account
		accountList, err = gClient.GetStatusFavouritedBy(opt.statusID)
		if opt.limit > 0 {
			accountList = accountList[:opt.limit]
		}
		obj = accountList
	case "delete":
		err = gClient.DeleteStatus(opt.statusID)
	case "boost":
		if opt.unset {
			err = gClient.UnreblogStatus(opt.statusID)
		} else {
			err = gClient.ReblogStatus(opt.statusID)
		}
	case "favourite":
		if opt.unset {
			err = gClient.UnfavouriteStatus(opt.statusID)
		} else {
			err = gClient.FavouriteStatus(opt.statusID)
		}
	case "post": // toot
		var s *madon.Status
		s, err = toot(strings.Join(args, " "))
		obj = s
	default:
		return errors.New("statusSubcommand: internal error")
	}

	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}
	if obj == nil {
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(obj, nil, "")
}
