package cmd

import (
	"fmt"
	"net/http"

	"github.com/mikeydub/go-gallery/indexer"
	"github.com/mikeydub/go-gallery/service/logger"
	sentryutil "github.com/mikeydub/go-gallery/service/sentry"
	"google.golang.org/appengine"

	"github.com/spf13/cobra"
)

var (
	port           uint64
	fromBlock      uint64
	toBlock        uint64
	enableRPC      bool
	quietLogs      bool
	serviceKeyFile string
)

var rootCmd = &cobra.Command{
	Use:   "indexer",
	Short: "Retrieve all NFTs that have ever been minted",
	Long: `An NFT indexer lovingly built by your friends at Gallery.
                Source code is available at https://github.com/gallery-so/go-gallery.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if toBlock < fromBlock {
			return fmt.Errorf("[from-block] must be less than [to-block]")
		}

		if !cmd.Flags().Lookup("to-block").Changed && !enableRPC {
			return fmt.Errorf("`flags in group [from-block, to-block] must all be set when [enable-rpc] is not set")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer sentryutil.RecoverAndRaise(nil)

		var from *uint64
		if cmd.Flags().Lookup("from-block").Changed {
			from = &fromBlock
		}

		var to *uint64
		if cmd.Flags().Lookup("to-block").Changed {
			to = &toBlock
		}

		indexer.Init(from, to, quietLogs, enableRPC)
		if appengine.IsAppEngine() {
			appengine.Main()
		} else {
			http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		}
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the indexer server",
	Run: func(cmd *cobra.Command, args []string) {
		defer sentryutil.RecoverAndRaise(nil)

		indexer.InitServer(serviceKeyFile, quietLogs, enableRPC)
		if appengine.IsAppEngine() {
			logger.For(nil).Info("Running in App Engine Mode")
			appengine.Main()
		} else {
			logger.For(nil).Info("Running in Default Mode")
			http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&enableRPC, "enable-rpc", "r", false, "always enable RPC calls")
	rootCmd.PersistentFlags().BoolVarP(&quietLogs, "quiet", "q", false, "hide debug logs")
	rootCmd.Flags().Uint64VarP(&port, "port", "p", 4000, "port to serve on")
	rootCmd.Flags().Uint64VarP(&fromBlock, "from-block", "f", 0, "first block to process")
	rootCmd.Flags().Uint64VarP(&toBlock, "to-block", "t", 0, "last block to process")
	rootCmd.MarkFlagsRequiredTogether("from-block", "to-block")

	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().Uint64VarP(&port, "port", "p", 6000, "port to serve on")
	serverCmd.Flags().StringVarP(&serviceKeyFile, "key-file", "f", "./_deploy/service-key-dev.json", "local service key file to use")
}

func Execute() {
	rootCmd.Execute()
}